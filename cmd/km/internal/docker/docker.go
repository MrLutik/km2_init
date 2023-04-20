package docker

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func PullImage(ctx context.Context, imageName string) error {

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Unable to create Docker client: %s", err)

	}
	options := types.ImagePullOptions{}
	reader, err := cli.ImagePull(ctx, imageName, options)
	if err != nil {
		return err
	}
	defer reader.Close()

	// TODO: Add buffer for reader. Pretify output
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}

	return nil
}
