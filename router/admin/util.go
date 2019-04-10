package admin

import (
	"github.com/level11consulting/ocelot/build/vcshandler/github"
	ocenet "github.com/shankj3/go-til/net"

	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"

	bb "github.com/level11consulting/ocelot/build/vcshandler/bitbucket"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"errors"
)

var unsupported = errors.New("currently only bitbucket is supported")

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss pb.GuideOcelotServer, config *pb.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)
	//hehe right now we only have bitbucket
	switch config.SubType {
	case pb.SubCredType_BITBUCKET:
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := bb.GetBitbucketHandler(config, bitbucketClient)
		go bbHandler.Walk() //spawning walk in a different thread because we don't want client to wait if there's a lot of repos/files to check
	case pb.SubCredType_GITHUB:
		cli, _, err := github.GetGithubClient(config)
		if err != nil {
			return err
		}
		go cli.Walk()
	default:
		return unsupported
	}

	config.Identifier = config.BuildIdentifier()
	//right now, we will always overwrite
	err := gos.RemoteConfig.AddCreds(gos.Storage, config, true)
	return err
}

func SetupRCCCredentials(remoteConf config.CVRemoteConfig, store storage.CredTable, config pb.OcyCredder) error {
	//right now, we will always overwrite
	err := remoteConf.AddCreds(store, config, true)
	return err
}

//RespWrap will wrap streaming messages in a LineResponse object to be sent by the server stream
func RespWrap(msg string) *pb.LineResponse {
	return &pb.LineResponse{OutputLine: msg}
}

// handleStorageError  will attempt to decipher if err is not found. if so, iwll set the appropriate grpc status code and return new grpc status error
func handleStorageError(err error) error {
	if _, ok := err.(*storage.ErrNotFound); ok {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
