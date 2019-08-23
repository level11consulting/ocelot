package runtime

import (
	"context"

	"github.com/level11consulting/orbitalci/build/vcshandler"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/pkg/errors"
)

// getAndSetHandler will use the accesstoken and vcstype to generate a handler without autorefresh capability and set it to (*launcher).handler field. if (*launcher).handler is already set, will do nothing
func (w *launcher) getAndSetHandler(ctx context.Context, accessToken string, vcsType pb.SubCredType) (err error) {
	if w.handler == nil {
		var handler models.VCSHandler
		handler, err = vcshandler.GetHandlerWithToken(ctx, accessToken, vcsType)
		if err != nil {
			return
		}
		w.handler = handler
	}
	return
}

// postFlight is what the launcher will do at the conclusion of the build, after all stages have run and everything is stored.
//   currently, if the build was signalled by a Pull Request, then a comment will be added to the PR in the subsequent VCS
func (w *launcher) postFlight(ctx context.Context, werk *pb.WerkerTask, failed bool) (err error) {
	if werk.SignaledBy == pb.SignaledBy_PULL_REQUEST {
		err = w.getAndSetHandler(ctx, werk.VcsToken, werk.VcsType)
		if err != nil {
			return
		}
		if err = w.handler.PostPRComment(werk.FullName, werk.PrData.PrId, werk.CheckoutHash, failed, werk.Id); err != nil {
			return
		}
		// only do approve for now, because decline is irreversible i guess??
		if !failed {
			if werk.PrData.Urls.Approve == "" {
				return errors.New("approve url is empty!!")
			}
			err = w.handler.GetClient().PostUrl(werk.PrData.Urls.Approve, "{}", nil)
			if err != nil {
				return
			}
		}
	}
	return nil
}
