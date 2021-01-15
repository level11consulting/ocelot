package admin

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shankj3/go-til/log"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

func (g *guideOcelotServer) DeletePollRepo(ctx context.Context, poll *pb.PollRequest) (*empty.Empty, error) {
	if poll.Account == "" || poll.Repo == "" {
		return nil, status.Error(codes.InvalidArgument, "account and repo are required fields")
	}
	log.Log().Info("received delete poll request for ", poll.Account, " ", poll.Repo)
	empti := &empty.Empty{}
	if err := g.Storage.DeletePoll(poll.Account, poll.Repo); err != nil {
		log.IncludeErrField(err).WithField("account", poll.Account).WithField("repo", poll.Repo).Error("couldn't delete poll")
	}
	log.Log().WithField("account", poll.Account).WithField("repo", poll.Repo).Info("successfully deleted poll in storage")
	if err := g.Producer.WriteProto(poll, "no_poll_please"); err != nil {
		log.IncludeErrField(err).Error("couldn't write to queue producer at no_poll_please")

		return empti, status.Error(codes.Unavailable, err.Error())
	}
	return empti, nil
}

// todo: add acct/repo action later
func (g *guideOcelotServer) ListPolledRepos(context.Context, *empty.Empty) (*pb.Polls, error) {
	polls, err := g.Storage.GetAllPolls()
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.Polls{Polls: polls}, nil
}
