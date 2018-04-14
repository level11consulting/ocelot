package main

// needs to:
// receive acct-repo as flag
// call bitbucket for changeset
// check if there have been updates, if there have:
//   - create build message from latest hash
//   - add build message to build topic
// 	 - update last_cron_time in db

import (
	"fmt"
	"os"
	"strings"

	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	pb "bitbucket.org/level11consulting/ocelot/old/protos"
	"bitbucket.org/level11consulting/ocelot/util/build"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/namsral/flag"
)

type changeSetConfig struct {
	RemoteConf   cred.CVRemoteConfig
	*deserialize.Deserializer
	OcyValidator   *build.OcelotValidator
	Producer       *nsqpb.PbProduce
	AcctRepo  	string
	Acct        string
	Repo        string
	Branches     []string
}

func configure() *changeSetConfig {
	var loglevel, consuladdr, acctRepo, branches string
	var consulport int
	flrg := flag.NewFlagSet("poller", flag.ExitOnError)
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.StringVar(&acctRepo, "acct-repo", "ERROR", "acct/repo to check changeset for")
	flrg.StringVar(&branches, "branches", "ERROR", "comma separated list of branches to check for changesets")
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.Parse(os.Args[1:])
	ocelog.InitializeLog(loglevel)
	ocelog.Log().Debug()
	rc, err := cred.GetInstance(consuladdr, consulport, "")
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get instance of remote config, exiting")
	}
	if acctRepo == "ERROR" || branches == "ERROR" {
		ocelog.Log().Fatal("-acct-repo and -branches is required")
	}
	branchList := strings.Split(branches, ",")
	conf := &changeSetConfig{RemoteConf: rc, AcctRepo: acctRepo, Branches:branchList, Deserializer: deserialize.New(), Producer: nsqpb.GetInitProducer(), OcyValidator: build.GetOcelotValidator()}
	acctrepolist := strings.Split(acctRepo, "/")
	if len(acctrepolist) != 2 {
		ocelog.Log().Fatal("-acct-repo must be in format <acct>/<repo>")
	}
	conf.Acct, conf.Repo = acctrepolist[0], acctrepolist[1]
	return conf
}


func main() {
	conf := configure()
	var bbHandler handler.VCSHandler
	var token string
	store, err := conf.RemoteConf.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("couldn't get storage")
	}
	defer store.Close()
	cfg, err := build.GetVcsCreds(store, conf.AcctRepo, conf.RemoteConf)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	bbHandler, token, err = handler.GetBitbucketClient(cfg)
	fmt.Println(token)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	_, lastHashes, err := store.GetLastData(conf.AcctRepo)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Error("couldn't get last cron time, setting last cron to 5 minutes ago")
	}
	// no matter what, we are inside the cron job, so we should be updating the db
	defer func(){
		if err = store.SetLastData(conf.Acct, conf.Repo, lastHashes); err != nil {
			ocelog.IncludeErrField(err).Error("unable to set last cron time")
			return
		}
		ocelog.Log().Info("successfully set last cron time")
		return
	}()

	for _, branch := range conf.Branches {
		lastHash, ok := lastHashes[branch]
		if !ok {
			ocelog.Log().Infof("no last hash found for branch %s in lash Hash map, so this branch will build no matter what", branch)
			lastHash = ""
		}
		newLastHash, err := searchBranchCommits(bbHandler, branch, conf, lastHash, store, token, &aWerkerTeller{})
		ocelog.Log().WithField("old last hash", lastHash).WithField("new last hash", newLastHash).Info("git hash data for poll")
		lastHashes[branch] = newLastHash
		if err != nil {
			ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("something went wrong")
		}
	}

}