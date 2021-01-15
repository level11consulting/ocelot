package build

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/common"
	util "github.com/level11consulting/ocelot/common/testutil"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"testing"
)

//// func CheckIfBuildDone(consulete *consul.Consulet, gitHash string) bool {
func Test_CheckIfBuildDone(t *testing.T) {
	hash := "sup"
	ctl := gomock.NewController(t)
	consu := consul.NewMockConsuletty(ctl)
	store := storage.NewMockBuildSum(ctl)
	store.EXPECT().AddSumStart(hash, "1", "2", "3", pb.SignaledBy_POLL, int64(1)).Return(int64(18), nil)
	store.EXPECT().UpdateSum(false, 10.7, int64(18)).Return(nil)
	testAddFullBuildSummary(hash, store, t)

	consu.EXPECT().GetKeyValue(common.MakeBuildMapPath(hash)).Return(nil, nil).Times(1)
	store.EXPECT().RetrieveLatestSum(hash).Return(&pb.BuildSummary{}, nil).Times(1)
	done := CheckIfBuildDone(consu, store, hash)
	if !done {
		t.Error(test.GenericStrFormatErrors("build done", true, done))
	}
	consu.EXPECT().GetKeyValue(common.MakeBuildMapPath("nerd")).Return(nil, nil).Times(1)
	store.EXPECT().RetrieveLatestSum("nerd").Return(nil, errors.New("np")).Times(1)
	done = CheckIfBuildDone(consu, store, "nerd")
	if done {
		t.Error(test.GenericStrFormatErrors("build done", false, done))
	}
}

func testAddFullBuildSummary(hash string, store storage.BuildSum, t *testing.T) {
	id, err := store.AddSumStart(hash, "1", "2", "3", pb.SignaledBy_POLL, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = store.UpdateSum(false, 10.7, id)
	if err != nil {
		t.Fatal(err)
	}
}

//
//// func GetBuildRuntime(consulete *consul.Consulet, gitHash string) (*BuildRuntime, error) {
func Test_GetBuildRuntime(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	werkerId := "werkerId"
	wsPort := "4030"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	serv.SetKV(t, fmt.Sprintf(common.WerkerGrpc, werkerId), []byte(grpcPort))
	serv.SetKV(t, fmt.Sprintf(common.WerkerWs, werkerId), []byte(wsPort))
	serv.SetKV(t, fmt.Sprintf(common.WerkerIp, werkerId), []byte(ip))
	serv.SetKV(t, fmt.Sprintf(common.BuildDockerUuid, werkerId, hash), []byte("dockerId"))
	serv.SetKV(t, fmt.Sprintf(common.WerkerBuildMap, hash), []byte(werkerId))
	brt, err := GetBuildRuntime(consu, hash)
	if err != nil {
		t.Fatal("unable to get build runtime, err: ", err.Error())
	}
	if len(brt) != 1 {
		t.Error(test.GenericStrFormatErrors("result length", 1, len(brt)))
	}

	for _, val := range brt {
		if val.Done != false {
			t.Error(test.GenericStrFormatErrors("done", false, val.Done))
		}
		if val.GrpcPort != grpcPort {
			t.Error(test.StrFormatErrors("grpc port", grpcPort, val.GrpcPort))
		}
		if val.Ip != ip {
			t.Error(test.StrFormatErrors("registered ip", ip, val.Ip))
		}
	}

}
