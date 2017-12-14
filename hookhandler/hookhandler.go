package hookhandler

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/admin/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"net/http"
)

type HookHandlerContext struct {
	RemoteConfig *cred.RemoteConfig
	Producer     *nsqpb.PbProduce
	Deserializer *deserialize.Deserializer
}

//TODO: look into all the branches that's listed inside of ocelot.yml and only build if event corresonds
//tODO: branch inside of ocelot.yml

//TODO: what data do we have to store/do we need to store?
// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func RepoPush(ctx *HookHandlerContext, w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}

	if err := ctx.Deserializer.JSONToProto(r.Body, repopush); err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not parse request body into proto.Message", err)
	}

	fullName := repopush.Repository.FullName
	hash := repopush.Push.Changes[0].New.Target.Hash
	acctName := repopush.Repository.Owner.Username
	buildConf, bbToken, err := GetBBConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", repopush.Repository.FullName)
		return
	}
	//TODO: need to check and make sure that New.Type == branch
	if validateBuild(buildConf, repopush.Push.Changes[0].New.Name) {
		tellWerker(ctx, buildConf, hash, fullName, acctName, bbToken)
	} else {
		//TODO: tell db we couldn't build
	}
}


// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func PullRequest(ctx *HookHandlerContext, w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := ctx.Deserializer.JSONToProto(r.Body, pr); err != nil {
		ocelog.IncludeErrField(err).Error("could not parse request body into pb.PullRequest")
		return
	}
	ocelog.Log().Debug(r.Body)
	fullName := pr.Pullrequest.Source.Repository.FullName
	hash := pr.Pullrequest.Source.Commit.Hash
	acctName := pr.Pullrequest.Source.Repository.Owner.Username

	buildConf, bbToken, err := GetBBConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", pr.Pullrequest.Source.Repository.FullName)
		return
	}

	if validateBuild(buildConf, "") {
		tellWerker(ctx, buildConf, hash, fullName, acctName, bbToken)
	} else {
		//TODO: tell db we couldn't build
	}
}

//before we build pipeline config for werker, validate and make sure this is good candidate
	// - check if commit branch matches with ocelot.yaml branch
	// - check if ocelot.yaml has at least one step called build
func validateBuild(buildConf *pb.BuildConfig, branch string) bool {
	_, ok := buildConf.Stages["build"]
	if !ok {
		return false
	}

	for _, buildBranch := range buildConf.Branches {
		if buildBranch == branch {
			return true
		}
	}
	return false
}

//TODO: this code needs to say X repo is now being tracked
//TODO: this code will also need to store status into db
//TODO: remove unused fields
func tellWerker(ctx *HookHandlerContext, buildConf *pb.BuildConfig, hash string, fullName string, acctName string, bbToken string) {
	// get one-time token use for access to vault
	token, err := ctx.RemoteConfig.Vault.CreateThrowawayToken()
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}

	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		BuildConf: buildConf,
		VcsToken: bbToken,
		VcsType: "bitbucket",
		FullName: fullName,
	}

	go ctx.Producer.WriteProto(werkerTask, "build")
}

func HandleBBEvent(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	handlerCtx := ctx.(*HookHandlerContext)

	switch r.Header.Get("X-Event-Key") {
	case "repo:push":
		RepoPush(handlerCtx, w, r)
	case "pullrequest:created",
		"pullrequest:updated":
		PullRequest(handlerCtx, w, r)
	default:
		ocelog.Log().Errorf("No support for Bitbucket event %s", r.Header.Get("X-Event-Key"))
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

// for testing
func getCredConfig() *models.Credentials {
	return &models.Credentials{
		ClientId:     "QEBYwP5cKAC3ykhau4",
		ClientSecret: "gKY2S3NGnFzJKBtUTGjQKc4UNvQqa2Vb",
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		AcctName:     "jessishank",
	}
}

//returns config if it exists, bitbucket token, and err
func GetBBConfig(ctx *HookHandlerContext, acctName string, repoFullName string, checkoutCommit string) (conf *pb.BuildConfig, token string, err error) {
	//cfg := getCredConfig()
	bbCreds, err := ctx.RemoteConfig.GetCredAt(cred.ConfigPath+"/bitbucket/"+acctName, false)
	cfg := bbCreds["bitbucket/"+acctName]
	bb := handler.Bitbucket{}

	bbClient := &ocenet.OAuthClient{}
	token, err = bbClient.Setup(cfg)

	bb.SetMeUp(cfg, bbClient)
	fileBitz, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		return
	}
	conf = &pb.BuildConfig{}
	if err != nil {
		return
	}
	if err = ctx.Deserializer.YAMLToStruct(fileBitz, conf); err != nil {
		return
	}
	return
}
