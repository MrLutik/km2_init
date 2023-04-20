package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
	"github.com/mrlutik/km2_init/km/internal/cosign"
)

func isValidSemVer(input string) error {
	_, err := semver.NewVersion(input)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	var (
		baseImageVer    string
		baseImageName   string
		sekaiContainer  bool
		interxContainer bool
	)

	ctx := context.Background()
	// Set latest version of the base-image
	flag.StringVar(&baseImageVer, "image", "v0.13.7", "Base-image version. Default: v0.13.7")

	// Set contatiners to launch
	// Binary will be from master branch aka latest
	flag.BoolVar(&sekaiContainer, "sekai", false, "Set to true to start container with sekai")
	flag.BoolVar(&interxContainer, "interx", false, "Set to true to start container with interx")

	flag.Parse()

	// Define the image you want to pull
	if err := isValidSemVer(baseImageVer); err != nil {
		fmt.Fprintln(os.Stderr, "semver is not valid")
		panic(err)
	}

	baseImageName = fmt.Sprintf("ghcr.io/kiracore/docker/base-image:%s", baseImageVer)

	// err := cosign.VerifyDockerImage(ctx, baseImageName, DockerImagePubKey)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }

	if verified, err := cosign.VerifyImageSignature(ctx, baseImageName, cosign.DockerImagePubKey); err != nil || verified != true {
		fmt.Fprintln(os.Stderr, "verification failed. err: ", err)
	}
	fmt.Fprintln(os.Stdout, "Image verified!")

}
