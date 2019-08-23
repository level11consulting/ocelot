package cleaner

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/shankj3/go-til/log"
)

type DockerCleaner struct{}

func (d *DockerCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	cli, err := client.NewEnvClient()
	defer cli.Close()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker client?? ")
		return err
	}

	if err = cli.ContainerKill(ctx, id, "SIGKILL"); err != nil {
		if err == context.Canceled && logout != nil {
			logout <- []byte("//////////REDRUM////////REDRUM////////REDRUM/////////")
		}
		log.IncludeErrField(err).WithField("containerId", id).Error("couldn't kill")
	} else {
		log.Log().WithField("dockerId", id).Info("killed container")
	}

	// even if ther is an error with containerKill, it might be from the container already exiting (ie bad ocelot.yml). so still try to remove.
	log.Log().WithField("dockerId", id).Info("removing")
	if err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		log.IncludeErrField(err).WithField("dockerId", id).Error("could not rm container")
		failedCleaning.WithLabelValues("docker").Inc()
		return err
	} else {
		log.Log().WithField("dockerId", id).Info("removed container")
	}
	return nil
}
