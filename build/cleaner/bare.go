package cleaner

import (
	"context"

	"github.com/shankj3/go-til/log"
)

type SSHCleaner struct {}

func (d *SSHCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	log.Log().Info("machine build with id ", id, "is finished. there is no cleanup for it.")
	//logout <- []byte("Finished Cleaning.")
	return nil
}