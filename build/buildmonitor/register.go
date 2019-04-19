package buildmonitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	consulkv "github.com/level11consulting/ocelot/server/config/consul"
	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
)

// Register will add all the appropriate build details that the admin needs to contact the werker for stream info
// will add:
// werkerLocation  = "ci/werker_location/%s" // %s is werker id
// ci/werker_location/<werkid> + werker_ip        = ip
// 		'' 			           + werker_grpc_port = grpcPort
// 		''				       + werker_ws_port   = wsPort
// 		''				       + tags		      = comma separated list of tags
// returns a generated uuid for the werker
func Register(consulete consul.Consuletty, ip, grpcPort, wsPort string, tags []string) (werkerId uuid.UUID, err error) {
	werkerId = uuid.New()
	strId := werkerId.String()
	if err = consulete.AddKeyValue(consulkv.MakeWerkerIpPath(strId), []byte(ip)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(consulkv.MakeWerkerGrpcPath(strId), []byte(grpcPort)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(consulkv.MakeWerkerWsPath(strId), []byte(wsPort)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(consulkv.MakeWerkerTagsPath(strId), []byte(strings.Join(tags, ","))); err != nil {
		return
	}
	return
}

//UnRegister deletes all values out of the kv store that are related to the werkerId's specific configuration
func UnRegister(consulete consul.Consuletty, werkerId string) error {
	err := consulete.RemoveValues(consulkv.MakeWerkerLocPath(werkerId))
	return err
}

// RegisterStartedBuild creates an entry in consul that maps the git hash to the werker's uuid so that clients can find the werker for live streaming
func RegisterStartedBuild(consulete consul.Consuletty, werkerId string, gitHash string) error {
	if err := consulete.AddKeyValue(consulkv.MakeBuildMapPath(gitHash), []byte(werkerId)); err != nil {
		return err
	}
	return nil
}

// RegisterBuild will add the mapping of docker uuid (or unique identifier, w/e) to the associated werkerId/commit build
func RegisterBuild(consulete consul.Consuletty, werkerId string, gitHash string, dockerUuid string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("docker_uuid", dockerUuid).Info("registering build")
	err := consulete.AddKeyValue(consulkv.MakeDockerUuidPath(werkerId, gitHash), []byte(dockerUuid))
	return err
}

// RegisterBuildSummaryId will associate the build_summary's database id number with the executing build
func RegisterBuildSummaryId(consulete consul.Consuletty, werkerId string, gitHash string, buildId int64) error {
	str := fmt.Sprintf("%d", buildId)
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildId", buildId).Info("registering build")
	err := consulete.AddKeyValue(consulkv.MakeBuildSummaryIdPath(werkerId, gitHash), []byte(str))
	return err
}

func RegisterBuildStage(consulete consul.Consuletty, werkerId string, gitHash string, buildStage string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildStage", buildStage).Info("registering build")
	err := consulete.AddKeyValue(consulkv.MakeBuildStagePath(werkerId, gitHash), []byte(buildStage))
	return err
}

func RegisterStageStartTime(consulete consul.Consuletty, werkerId string, gitHash string, start time.Time) error {
	str := fmt.Sprintf("%d", start.Unix()) // todo: figure out a better way to do this conversion using bit shifting or something because i know this isnt the "right" way to do it
	err := consulete.AddKeyValue(consulkv.MakeBuildStartpath(werkerId, gitHash), []byte(str))
	return err
}
