package build_signaler

import (
	"errors"
	"fmt"
	"time"

	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

func storeStageToDb(store storage.BuildStage, stageResult *models.StageResult) error {
	if err := store.AddStageDetail(stageResult); err != nil {
		log.IncludeErrField(err).Error("unable to store hookhandler stage details to db")
		return err
	}
	return nil
}

func storeQueued(store storage.BuildSum, id int64) error {
	err := store.SetQueueTime(id)
	if err != nil {
		log.IncludeErrField(err).Error("unable to update queue time in build summary table")
	}
	return err
}

func storeSummaryToDb(store storage.BuildSum, hash, repo, branch, account string, by pb.SignaledBy, credId int64) (int64, error) {
	id, err := store.AddSumStart(hash, account, repo, branch, by, credId)
	if err != nil {
		log.IncludeErrField(err).Error("unable to store summary details to db")
		return 0, err
	}
	return id, nil
}

//todo: pull out check for vcsHandler == nil logic, then this can be just GetConfig()
// todo (cont): write something in remote to switch between subtypes to instantiate the correct VCSHandler implementation
//GetConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err.
//If a VcsHandler is passed, this method will use the existing handler to retrieve the bb config. In that case,
//***IT WILL NOT RETURN A VALID TOKEN FOR YOU - ONLY BUILD CONFIG***
func GetConfig(repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler models.VCSHandler) (*pb.BuildConfig, error) {
	//var bbHandler remote.VCSHandler
	//var token string

	if vcsHandler == nil {
		return nil, errors.New("vcs handler cannot be nul")
	}

	fileBytz, err := vcsHandler.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, err
	}

	conf, err := CheckForBuildFile(fileBytz, deserializer)
	return conf, err
}

//CheckForBuildFile will try to retrieve an ocelot.yaml file for a repository and return the protobuf message
func CheckForBuildFile(buildFile []byte, deserializer *deserialize.Deserializer) (*pb.BuildConfig, error) {
	conf := &pb.BuildConfig{}
	fmt.Println(string(buildFile))
	if err := deserializer.YAMLToStruct(buildFile, conf); err != nil {
		if err != ocenet.FileNotFound {
			log.IncludeErrField(err).Error("unable to get build conf")
			return conf, err
		}
		log.Log().Debugf("no ocelot yml found")
		return conf, err
	}
	return conf, nil
}

func PopulateStageResult(sr *models.StageResult, status int, lastMsg, errMsg string) {
	sr.Messages = append(sr.Messages, lastMsg)
	sr.Status = status
	sr.Error = errMsg
	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
}

// all this moved to build_signaler.go
func getSignalerStageResult(id int64) *models.StageResult {
	start := time.Now()
	return &models.StageResult{
		Messages:      []string{},
		BuildId:       id,
		Stage:         models.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99,
	}
}
