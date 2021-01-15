package bitbucket

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models"
	pbb "github.com/level11consulting/ocelot/models/bitbucket/pb"
	"github.com/level11consulting/ocelot/models/pb"
)

const DefaultRepoBaseURL = "https://api.bitbucket.org/2.0/repositories/%v"

const TokenUrl = "https://bitbucket.org/site/oauth2/access_token"

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetBitbucketClient(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	cfg.TokenURL = TokenUrl
	bbClient := &ocenet.OAuthClient{}
	token, err := bbClient.Setup(cfg)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + cfg.AcctName + ".  Error: " + err.Error())
	}
	bb := GetBitbucketHandler(cfg, bbClient)
	return bb, token, nil
}

func GetBitbucketFromHttpClient(cli *http.Client) models.VCSHandler {
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	bb := &Bitbucket{
		Client:        &ocenet.OAuthClient{AuthClient: cli, Unmarshaler: unmarshaler},
		Unmarshaler:   unmarshaler,
		Marshaler:     jsonpb.Marshaler{},
		isInitialized: true,
	}
	return bb
}

//TODO: callback url is set as env. variable on admin, or passed in via command line
//GetBitbucketHandler returns a Bitbucket handler referenced by VCSHandler interface
func GetBitbucketHandler(adminConfig *pb.VCSCreds, client ocenet.HttpClient) models.VCSHandler {
	bb := &Bitbucket{
		Client:        client,
		Marshaler:     jsonpb.Marshaler{},
		Unmarshaler:   jsonpb.Unmarshaler{AllowUnknownFields: true},
		credConfig:    adminConfig,
		isInitialized: true,
	}
	return bb
}

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	CallbackURL string
	RepoBaseURL string
	Client      ocenet.HttpClient
	Marshaler   jsonpb.Marshaler
	Unmarshaler jsonpb.Unmarshaler

	credConfig    *pb.VCSCreds
	isInitialized bool
}

func (bb *Bitbucket) GetVcsType() pb.SubCredType {
	return pb.SubCredType_BITBUCKET
}

func (bb *Bitbucket) GetClient() ocenet.HttpClient {
	return bb.Client
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (bb *Bitbucket) Walk() error {
	if !bb.isInitialized {
		return errors.New("client has not yet been initialized, please call SetMeUp() before walking")
	}
	return bb.recurseOverRepos(fmt.Sprintf(bb.GetBaseURL(), bb.credConfig.AcctName))
}

// Get File in repo at a certain commit.
// filepath: string filepath relative to root of repo
// fullRepoName: string account_name/repo_name as it is returned in the Bitbucket api Repo Source `full_name`
// commitHash: string git hash for revision number
func (bb *Bitbucket) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	ocelog.Log().Debug("inside GetFile")
	path := fmt.Sprintf("%s/src/%s/%s", fullRepoName, commitHash, filePath)
	bytez, err = bb.Client.GetUrlRawData(fmt.Sprintf(bb.GetBaseURL(), path))
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("GetFile").Inc()
		ocelog.IncludeErrField(err).Error()
		return
	}
	return
}

func translateBbCommit(commit *pbb.Commit) *pb.Commit {
	return &pb.Commit{
		Hash: commit.Hash,
		Message: commit.Message,
		Date: commit.Date,
		Author: &pb.User{UserName: commit.Author.User.Username},
	}
}

//GetAllCommits /2.0/repositories/{username}/{repo_slug}/commits
func (bb *Bitbucket) GetAllCommits(acctRepo string, branch string) ([]*pb.Commit, error) {
	commits := &pbb.Commits{}
	err := bb.Client.GetUrl(fmt.Sprintf(bb.GetBaseURL(), acctRepo)+"/commits/"+branch, commits)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("GetAllCommits").Inc()
	}
	var translatedCommits []*pb.Commit
	for _, commit := range commits.Values {
		translatedCommits = append(translatedCommits, translateBbCommit(commit))
	}
	return translatedCommits, err
}

//GetCommitLog will return a list of Commits, starting with the most recent and ending at the lastHash value.
// If the lastHash commit value is never found, will return an error.
func (bb *Bitbucket) GetCommitLog(acctRepo, branch, lastHash string) ([]*pb.Commit, error) {
	var commits []*pb.Commit
	if lastHash == "" {
		return commits, nil
	}
	var foundLast bool
	urrl := fmt.Sprintf(bb.GetBaseURL(), acctRepo) + "/commits/" + branch
	for {
		if urrl == "" || foundLast == true {
			break
		}
		commitz := &pbb.Commits{}
		err := bb.Client.GetUrl(urrl, commitz)
		if err != nil {
			failedBBRemoteCalls.WithLabelValues("GetCommitLog").Inc()
			return commits, err
		}
		for _, commit := range commitz.Values {
			commits = append(commits, &pb.Commit{Hash: commit.Hash, Message: commit.Message, Date: commit.Date})
			if commit.Hash == lastHash {
				foundLast = true
				break
			}
		}
		urrl = commitz.GetNext()
	}
	var err error
	if !foundLast {
		err = models.Commit404(lastHash, acctRepo, branch)
	}
	return commits, err
}

func (bb *Bitbucket) GetRepoLinks(acctRepo string) (*pb.Links, error) {
	repoVal := &pbb.PaginatedRepository_RepositoryValues{}
	err := bb.Client.GetUrl(fmt.Sprintf(DefaultRepoBaseURL, acctRepo), repoVal)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("GetRepoDetail").Inc()
		return nil, err
	}
	if repoVal.Type == "error" {
		return nil, errors.New(fmt.Sprintf("could not get repository detail at %s", acctRepo))
	}

	links := &pb.Links{
		Commits: repoVal.Links.Commits.Href,
		Branches: repoVal.Links.Branches.Href,
		Tags: repoVal.Links.Tags.Href,
		Hooks: repoVal.Links.Hooks.Href,
		Pullrequests: repoVal.Links.Pullrequests.Href,
	}
	return links, nil
}

func (bb *Bitbucket) GetBranchLastCommitData(acctRepo, branch string) (hist *pb.BranchHistory, err error) {
	path := fmt.Sprintf("%s/refs/branches/%s", acctRepo, branch)
	urrl := fmt.Sprintf(bb.GetBaseURL(), path)
	var resp *http.Response
	resp, err = bb.Client.GetUrlResponse(urrl)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("GetBranchLastCommitData").Inc()
		return nil, err
	}
	defer resp.Body.Close()
	// status code handling using bitbucket API spec
	//   https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Busername%7D/%7Brepo_slug%7D/refs/branches/%7Bname%7D
	switch resp.StatusCode {
	case http.StatusNotFound:
		err = models.Branch404(branch, acctRepo)
	case http.StatusForbidden:
		err = errors.New(fmt.Sprintf("Repo %s (with branch %s) is private and these credentials are not authorized for access", acctRepo, branch))
	case http.StatusOK:
		bbBranch := &pbb.Branch{}
		reader := bufio.NewReader(resp.Body)
		if err = bb.Unmarshaler.Unmarshal(reader, bbBranch); err != nil {
			ocelog.IncludeErrField(err).Error("failed to parse response from ", urrl)
			return
		}
		hist = &pb.BranchHistory{Branch: branch, Hash: bbBranch.GetTarget().GetHash(), LastCommitTime: bbBranch.GetTarget().GetDate()}
		err = nil
	}
	return
}

func (bb *Bitbucket) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	var branchHistory []*pb.BranchHistory
	var nextUrl string
	path := fmt.Sprintf("%s/refs/branches", acctRepo)
	nextUrl = fmt.Sprintf(bb.GetBaseURL(), path)
	for {
		branches := &pbb.PaginatedRepoBranches{}
		err := bb.Client.GetUrl(nextUrl, branches)
		if err != nil {
			failedBBRemoteCalls.WithLabelValues("GetAllBranchesLastCommitData").Inc()
			return nil, err
		}
		for _, branch := range branches.GetValues() {
			branchHistory = append(branchHistory, &pb.BranchHistory{Branch: branch.Name, Hash: branch.Target.GetHash(), LastCommitTime: branch.Target.GetDate()})
		}
		nextUrl = branches.GetNext()
		if nextUrl == "" {
			break
		}
	}
	return branchHistory, nil
}

//CreateWebhook will create webhook at specified webhook url
func (bb *Bitbucket) CreateWebhook(webhookURL string) error {
	if bb.CallbackURL == "" {
		return models.NoCallbackURL(pb.SubCredType_BITBUCKET)
	}
	if !bb.FindWebhooks(webhookURL) {
		//create webhook if one does not already exist
		newWebhook := &pbb.CreateWebhook{
			Description: "marianne did this",
			Active:      true,
			Url:         bb.GetCallbackURL(),
			Events:      common.BitbucketEvents,
		}
		webhookStr, err := bb.Marshaler.MarshalToString(newWebhook)
		if err != nil {
			ocelog.IncludeErrField(err).Fatal("failed to convert webhook to json string")
			return err
		}
		err = bb.Client.PostUrl(webhookURL, webhookStr, nil)
		if err != nil {
			failedBBRemoteCalls.WithLabelValues("CreateWebhook").Inc()
			return err
		}
		ocelog.Log().Debug("subscribed to webhook for ", webhookURL)
	}
	return nil
}

//GetCallbackURL is a getter for retrieving callbackURL for bitbucket webhooks
func (bb *Bitbucket) GetCallbackURL() string {
	return bb.CallbackURL + "/" + strings.ToLower(bb.GetVcsType().String())
}

//SetCallbackURL sets callback urls to be used for webhooks
func (bb *Bitbucket) SetCallbackURL(callbackURL string) {
	bb.CallbackURL = callbackURL
}

func (bb *Bitbucket) SetBaseURL(baseURL string) {
	bb.RepoBaseURL = baseURL
}

func (bb *Bitbucket) GetBaseURL() string {
	if len(bb.RepoBaseURL) > 0 {
		return bb.RepoBaseURL
	}
	return DefaultRepoBaseURL
}

//recursively iterates over all repositories and creates webhook
func (bb *Bitbucket) recurseOverRepos(repoUrl string) error {
	if repoUrl == "" {
		return nil
	}
	repositories := &pbb.PaginatedRepository{}
	//todo: error pages from bitbucket??? these need to bubble up to client
	err := bb.Client.GetUrl(repoUrl, repositories)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("recurseOverRepos").Inc()
		return err
	}

	for _, v := range repositories.GetValues() {
		ocelog.Log().Debug(fmt.Sprintf("found repo %v", v.GetFullName()))
		err = bb.recurseOverFiles(v.GetLinks().GetSource().GetHref(), v.GetLinks().GetHooks().GetHref())
		if err != nil {
			return err
		}
	}
	return bb.recurseOverRepos(repositories.GetNext())
}

//recursively iterates over all source files trying to find build file
func (bb Bitbucket) recurseOverFiles(sourceFileUrl string, webhookUrl string) error {
	if sourceFileUrl == "" {
		return nil
	}
	repositories := &pbb.PaginatedRootDirs{}
	err := bb.Client.GetUrl(sourceFileUrl, repositories)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("recurseOverFiles").Inc()
		return err
	}
	for _, v := range repositories.GetValues() {
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 && v.GetPath() == common.BuildFileName {
			ocelog.Log().Debug("holy crap we actually an ocelot.yml file")
			err = bb.CreateWebhook(webhookUrl)
			if err != nil {
				return err
			}
		}
	}
	return bb.recurseOverFiles(repositories.GetNext(), webhookUrl)
}

//recursively iterates over all webhooks and returns true (matches our callback urls) if one already exists
func (bb *Bitbucket) FindWebhooks(getWebhookURL string) bool {
	if getWebhookURL == "" {
		return false
	}
	webhooks := &pbb.GetWebhooks{}
	err := bb.Client.GetUrl(getWebhookURL, webhooks)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("FindWebhooks").Inc()
	}

	for _, wh := range webhooks.GetValues() {
		if wh.GetUrl() == bb.GetCallbackURL() {
			return true
		}
	}

	return bb.FindWebhooks(webhooks.GetNext())
}

func (bb *Bitbucket) GetPRCommits(url string) ([]*pb.Commit, error) {
	var commits []*pb.Commit
	for {
		if url == "" {
			break
		}
		commitz := &pbb.Commits{}
		err := bb.Client.GetUrl(url, commitz)
		if err != nil {
			failedBBRemoteCalls.WithLabelValues("GetPRCommits").Inc()
			return commits, err
		}
		for _, commit := range commitz.Values {
			commits = append(commits, &pb.Commit{Hash: commit.Hash, Message: commit.Message, Date: commit.Date, Author: &pb.User{UserName: commit.Author.User.Username}})
		}
		url = commitz.GetNext()
	}
	return commits, nil
}

func (bb *Bitbucket) PostPRComment(acctRepo, prId, hash string, fail bool, buildId int64) error {
	path := fmt.Sprintf("%s/pullrequests/%s/comments", acctRepo, prId)
	urll := fmt.Sprintf(bb.GetBaseURL(), path)
	var status string
	switch fail {
	case true:
		status = "FAILED"
	case false:
		status = "PASSED"
	}
	content := fmt.Sprintf("Ocelot build has **%s** for commit **%s**.\n\nRun `ocelot status -build-id %d` for detailed stage status, and `ocelot run -build-id %d` for complete build logs.", status, hash, buildId, buildId)
	body := map[string]map[string]string{
		"content": {
			"raw": content,
			//"markup": "markdown",
		},
	}
	bodybytes, _ := json.Marshal(body)
	resp, err := bb.Client.GetAuthClient().Post(urll, "application/json", bytes.NewReader(bodybytes))
	defer resp.Body.Close()
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("PostPRComment").Inc()
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(fmt.Sprintf("got a non-ok exit code of %d, body is: %s", resp.StatusCode, string(body)))
		return err
	}
	return err
}

//GetChangedFiles will get the list of files that have changed between commits. If earliestHash is not passed,
// then the diff list will be off of just the changed files in the latestHash. If earliesthash is passed, then it will
// return the changeset similar to git diff --name-only <latestHash>..<earliestHash>
func (bb *Bitbucket) GetChangedFiles(acctRepo, latestHash, earliestHash string) (changedFiles []string, err error) {
	changedFileSet := map[string]bool{}
	// https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/diffstat/d222fa2..e174964

	var diffStatPath string
	if earliestHash != "" {
		diffStatPath = fmt.Sprintf("%s..%s", latestHash, earliestHash)
	} else {
		diffStatPath = latestHash
	}
	path := fmt.Sprintf("%s/diffstat/%s", acctRepo, diffStatPath)
	urll := fmt.Sprintf(bb.GetBaseURL(), path)
	for {
		if urll == "" {
			break
		}
		diff := &pbb.FullDiff{}
		err := bb.Client.GetUrl(urll, diff)
		if err != nil {
			failedBBRemoteCalls.WithLabelValues("GetChangedFiles").Inc()
			return changedFiles, err
		}
		for _, diffstat := range diff.Values {
			if diffstat.New != nil {
				changedFileSet[diffstat.New.Path] = true
			}
			if diffstat.Old != nil {
				changedFileSet[diffstat.Old.Path] = true
			}
		}
		urll = diff.GetNext()
	}
	changedFiles = common.GetMapStringKeys(changedFileSet)
	return
}

func (bb *Bitbucket) GetCommit(acctRepo, hash string) (*pb.Commit, error) {
	path := fmt.Sprintf("%s/commit/%s", acctRepo, hash)
	urll := fmt.Sprintf(bb.GetBaseURL(), path)
	commit := &pbb.Commit{}
	err := bb.Client.GetUrl(urll, commit)
	if err != nil {
		failedBBRemoteCalls.WithLabelValues("GetCommit").Inc()
		return nil, err
	}
	var author *pb.User
	if commit.GetAuthor() != nil {
		if user := commit.Author.GetUser(); user != nil {
			author = &pb.User{UserName: user.Username, DisplayName: user.DisplayName}
		}
	}
	translatedCommit := &pb.Commit{Message: commit.Message, Hash: commit.Hash, Date: commit.Date, Author: author}
	return translatedCommit, nil

}
