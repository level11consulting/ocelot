package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/google/go-github/v19/github"
	"github.com/pkg/errors"

	"github.com/level11consulting/ocelot/build/vcshandler/config"
	"github.com/level11consulting/ocelot/models"
	gpb "github.com/level11consulting/ocelot/models/github/pb"
	"github.com/level11consulting/ocelot/models/pb"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

const DefaultBaseURL = "https://api.github.com/%s"

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetGithubClient(creds *pb.VCSCreds) (models.VCSHandler, string, error) {
	client := &ocenet.OAuthClient{}
	token, err := client.SetupStaticToken(creds)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + creds.AcctName + ".  Error: " + err.Error())
	}
	gh := GetGithubHandler(creds, client)
	return gh, token, nil
}

func GetGithubFromHttpClient(cli *http.Client) models.VCSHandler {
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	return &githubVCS{
		Unmarshaler: unmarshaler,
		Client:      &ocenet.OAuthClient{AuthClient: cli, Unmarshaler: unmarshaler},
		Marshaler:   jsonpb.Marshaler{},
	}
}

func GetGithubHandler(cred *pb.VCSCreds, cli ocenet.HttpClient) *githubVCS {
	return &githubVCS{
		Client:      cli,
		Marshaler:   jsonpb.Marshaler{},
		Unmarshaler: jsonpb.Unmarshaler{AllowUnknownFields: true},
		credConfig:  cred,
		ghClient:    github.NewClient(cli.GetAuthClient()),
		ctx:         context.Background(),
	}
}

type githubVCS struct {
	CallbackURL string
	Client      ocenet.HttpClient
	ghClient    *github.Client
	ctx         context.Context
	Marshaler   jsonpb.Marshaler
	Unmarshaler jsonpb.Unmarshaler
	credConfig  *pb.VCSCreds
	baseUrl     string
	// for testing
	setCommentId int64
}

func (gh *githubVCS) GetVcsType() pb.SubCredType {
	return pb.SubCredType_GITHUB
}

func (gh *githubVCS) GetCallbackURL() string {
	return gh.CallbackURL + "/" + strings.ToLower(gh.GetVcsType().String())
}

func (gh *githubVCS) SetCallbackURL(cbUrl string) {
	gh.CallbackURL = cbUrl
}

func (gh *githubVCS) GetClient() ocenet.HttpClient {
	return gh.Client
}

func (gh *githubVCS) GetBaseURL() string {
	if gh.baseUrl != "" {
		return gh.baseUrl
	}
	return DefaultBaseURL
}

func (gh *githubVCS) SetBaseURL(baseUrl string) {
	gh.baseUrl = baseUrl
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (gh *githubVCS) Walk() error {
	return gh.recurseOverRepos(0)
}

// recurseOverRepos will iterate over every repository checking for an ocelot.yml
// if one is found, then a webhook will be created
func (gh *githubVCS) recurseOverRepos(pageNum int) error {

	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 50, Page: pageNum},
	}
	ocelog.Log().Info("checking all the repos")
	repos, resp, err := gh.ghClient.Repositories.List(gh.ctx, "", opts)
	if err != nil {
		return errors.Wrap(err, "unable to list all repos")
	}
	resp.Body.Close()
	for _, repo := range repos {
		statusCode, erro := gh.checkForOcelotFile(repo.GetContentsURL())
		if erro != nil {
			return erro
		}
		if statusCode == http.StatusOK {
			ocelog.Log().Infof("%s has an ocelot.yml file!", repo.GetName())
			if err = gh.CreateWebhook(repo.GetHooksURL()); err != nil {
				return errors.Wrap(err, "unable to create webhook")
			}
		}
	}
	if resp.NextPage == 0 {
		return nil
	}
	return gh.recurseOverRepos(resp.NextPage)
}

//checkForOcelotFile will attempt to retrieve the http status of a request for a file at the path `ocelot.yml`. It will
// return the status code which can then be checked for if the file exists
func (gh *githubVCS) checkForOcelotFile(contentsUrl string) (int, error) {
	resp, err := gh.Client.GetUrlResponse(getUrlForFileFromContentsUrl(contentsUrl, config.BuildFileName))
	if err != nil {
		return 0, errors.Wrap(err, "unable to see if ocelot.yml exists")
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (gh *githubVCS) CreateWebhook(hookUrl string) error {
	if gh.CallbackURL == "" {
		return models.NoCallbackURL(pb.SubCredType_GITHUB)
	}
	// create it, if it already exists it'll return a 422
	hookReq := &gpb.Hook{
		Active: true,
		Events: []string{"push", "pull_request"},
		Config: &gpb.Hook_Config{Url: gh.GetCallbackURL(), ContentType: "json"},
	}
	bits, err := json.Marshal(hookReq)
	if err != nil {
		return errors.Wrap(err, "couldn't marshal hook request to json")
	}
	var resp *http.Response
	resp, err = gh.Client.GetAuthClient().Post(hookUrl, "application/json", bytes.NewReader(bits))
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("CreateWebhook").Inc()
		return errors.Wrap(err, "unable to complete webhook create")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		ocelog.Log().Infof("successfully created webhook with %s", hookUrl)
		return nil
	}
	ghErr := &gpb.Error{}
	_ = jsonpb.Unmarshal(resp.Body, ghErr)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		if ghErr.Message == "Validation Failed" {
			return nil
		}
	}
	err = errors.New(resp.Status + ": " + ghErr.Message)
	ocelog.IncludeErrField(err).Error("unable to create webhook!")
	return err
}

// GetFile will retrieve a file at {filePath} from account/repository specified by {fullRepoName} at the commitHash using the github api
func (gh *githubVCS) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	logWithFields := ocelog.Log().WithField("filePath", filePath).WithField("fullRepoName", fullRepoName).WithField("hash", commitHash)
	logWithFields.Debug("getting file ")
	acct, repo := splitAcctRepo(fullRepoName)
	getOpts := &github.RepositoryContentGetOptions{Ref: commitHash}
	var contents io.ReadCloser
	contents, err = gh.ghClient.Repositories.DownloadContents(gh.ctx, acct, repo, filePath, getOpts)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetFile").Inc()
		logWithFields.WithField("err", err.Error()).Error("cannot get file contents")
		err = errors.Wrap(err, "unable to get file contents")
		return
	}
	defer contents.Close()
	bytez, err = ioutil.ReadAll(contents)
	return
}

func (gh *githubVCS) GetRepoLinks(acctRepo string) (*pb.Links, error) {
	logWithFields := ocelog.Log().WithField("acctRepo", acctRepo)
	logWithFields.Debug("getting repo links")
	acct, repo := splitAcctRepo(acctRepo)
	repository, resp, err := gh.ghClient.Repositories.Get(gh.ctx, acct, repo)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetRepoLinks").Inc()
		ocelog.IncludeErrField(err).Error("cannot get repo links")
		return nil, errors.Wrap(err, "unable to get repository links")
	}
	defer resp.Body.Close()
	links := &pb.Links{
		Commits:      repository.GetCommitsURL(),
		Branches:     repository.GetBranchesURL(),
		Tags:         repository.GetTagsURL(),
		Hooks:        repository.GetHooksURL(),
		Pullrequests: repository.GetPullsURL(),
	}
	logWithFields.Debug("got repo links")
	return links, nil
}

func (gh *githubVCS) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	var branchesHistory []*pb.BranchHistory
	acct, repo := splitAcctRepo(acctRepo)
	opts := &github.ListOptions{PerPage: 50}
	for {
		branches, resp, err := gh.ghClient.Repositories.ListBranches(gh.ctx, acct, repo, opts)
		if err != nil {
			failedGHRemoteCalls.WithLabelValues("GetAllBrancheseLastCommitData").Inc()
			ocelog.IncludeErrField(err).WithField("acctRepo", acctRepo).Error("cannot get branch data")
			return nil, err
		}
		for _, branch := range branches {
			branchesHistory = append(branchesHistory, translateToBranchHistory(branch))
		}
		resp.Body.Close()
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return branchesHistory, nil
}

func (gh *githubVCS) GetBranchLastCommitData(acctRepo, branch string) (history *pb.BranchHistory, err error) {
	logWithFields := ocelog.Log().WithField("acctRepo", acctRepo).WithField("branch", branch)
	logWithFields.Debug("getting branch last commit data")
	acct, repo := splitAcctRepo(acctRepo)
	brch, resp, err := gh.ghClient.Repositories.GetBranch(gh.ctx, acct, repo, branch)
	if err != nil {
		logWithFields.WithField("err", err.Error()).Error("unable to get last commit data")
		failedGHRemoteCalls.WithLabelValues("GetBranchLastCommitData").Inc()
		err = errors.Wrap(err, "unable to get last commit data")
		return
	}
	logWithFields.Debug("successfully got branch last commit data")
	defer resp.Body.Close()
	history = translateToBranchHistory(brch)
	return
}

func (gh *githubVCS) GetCommitLog(acctRepo string, branch string, lastHash string) (commits []*pb.Commit, err error) {
	logWithFields := ocelog.Log().WithField("acctRepo", acctRepo).WithField("branch", branch).WithField("lastHash", lastHash)
	logWithFields.Debug("getting commit log")
	acct, repo := splitAcctRepo(acctRepo)
	opt := &github.CommitsListOptions{
		SHA: branch,
		ListOptions: github.ListOptions{
			PerPage: 40,
		},
	}
	for {
		ghCommits, resp, err := gh.ghClient.Repositories.ListCommits(gh.ctx, acct, repo, opt)
		if err != nil {
			failedGHRemoteCalls.WithLabelValues("GetCommitLog").Inc()
			logWithFields.WithField("err", err).Error("unable to get list of commits!")
			return nil, errors.Wrap(err, "unable to get list of commits")
		}
		resp.Body.Close()
		for _, ghCommit := range ghCommits {
			commits = append(commits, translateToCommit(ghCommit))
			if ghCommit.GetSHA() == lastHash {
				goto RETURN
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
RETURN:
	logWithFields.Debug("got commit log!")
	return
}

func (gh *githubVCS) PostPRComment(acctRepo, prId, hash string, failed bool, buildId int64) error {
	logWithField := ocelog.Log().WithField("acctRepo", acctRepo).WithField("prId", prId).WithField("hash", hash).WithField("failed", failed).WithField("buildId", buildId)
	logWithField.Debug("going to post pr comment")
	acct, repo := splitAcctRepo(acctRepo)
	var status string
	switch failed {
	case true:
		status = "FAILED"
	case false:
		status = "PASSED"
	}
	content := fmt.Sprintf("Ocelot build has **%s** for commit **%s**.\n\nRun `ocelot status -build-id %d` for detailed stage status, and `ocelot run -build-id %d` for complete build logs.", status, hash, buildId, buildId)
	prIdInt, err := strconv.Atoi(prId)
	if err != nil {
		return errors.Wrap(err, "invalid pr id")
	}
	comment := &github.IssueComment{Body: &content}
	cmt, resp, err := gh.ghClient.Issues.CreateComment(gh.ctx, acct, repo, prIdInt, comment)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("PostPRComment").Inc()
		logWithField.WithField("err", err.Error()).Error("unable to create pr comment")
		return errors.Wrap(err, "unable to create a pr comment")
	}
	resp.Body.Close()
	gh.setCommentId = cmt.GetID()
	logWithField.Debug("successfully posted pr comment")
	return nil
}

// for testing
func (gh *githubVCS) deleteIssueComment(account, repo string, commentId int64) error {
	resp, err := gh.ghClient.Issues.DeleteComment(gh.ctx, account, repo, commentId)
	if err != nil {
		ocelog.IncludeErrField(err).Error("bad delete")
		return err
	}
	resp.Body.Close()
	return nil
}

// for testing
func (gh *githubVCS) getIssueComment(account, repo string, commentID int64) error {
	comment, resp, err := gh.ghClient.Issues.GetComment(gh.ctx, account, repo, commentID)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if comment == nil {
		return errors.New("not found")
	}
	return nil
}

func (gh *githubVCS) GetChangedFiles(acctRepo, latesthash, earliestHash string) ([]string, error) {
	if earliestHash == "" {
		earliestHash = latesthash + "~1"
	}
	logWithFields := ocelog.Log().WithField("acctRepo", acctRepo).WithField("latestHash", latesthash).WithField("earliestHash", earliestHash)
	logWithFields.Debug("getting changed files")
	var changedFiles []string
	//GET /repos/:owner/:repo/compare/:base...:head
	acct, repo := splitAcctRepo(acctRepo)
	compare, resp, err := gh.ghClient.Repositories.CompareCommits(gh.ctx, acct, repo, earliestHash, latesthash)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetChangedFiles").Inc()
		logWithFields.WithField("err", err.Error()).Error("unable to get changed files")
		return nil, errors.Wrap(err, "unable to get changed files")
	}
	resp.Body.Close()
	for _, file := range compare.Files {
		changedFiles = append(changedFiles, file.GetFilename())
	}
	logWithFields.Debug("successfully got changed files!")
	return changedFiles, nil
}

func (gh *githubVCS) GetCommit(acctRepo, hash string) (*pb.Commit, error) {
	acct, repo := splitAcctRepo(acctRepo)
	commit, resp, err := gh.ghClient.Repositories.GetCommit(gh.ctx, acct, repo, hash)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetCommit").Inc()
		ocelog.IncludeErrField(err).Error("cannot get commit")
		return nil, errors.Wrap(err, "cannot get commit")
	}
	resp.Body.Close()
	return translateToCommit(commit), nil
}
