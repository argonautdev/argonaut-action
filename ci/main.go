package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"

	"dagger.io/dagger"
	"go.uber.org/zap"
)

func main() {
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {

	fmt.Println("[fmt]build process started")
	zap.S().Info("[z]build process started")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	buildId := os.Args[0]
	if buildId == "" {
		return errors.New("argonaut build identifier is missing")
	}

	userRepoLoc := os.Args[1]
	if userRepoLoc == "" {
		return errors.New("user repo location missing")
	}

	argKey := os.Getenv("ARG_KEY")
	argSecret := os.Getenv("ARG_SECRET")

	if argKey == "" || argSecret == "" {
		return errors.New("access to argonaut server is not configured")
	}

	mc, err := GetMidgardClient(argKey, argSecret)
	if err != nil {
		return err
	}

	buildInfo, err := mc.FetchBuildInfo(buildId)
	if err != nil {
		return err
	}

	//cache := client.CacheVolume("argonaut")

	contextDir := client.Host().Directory(userRepoLoc)

	ref, err := client.Container().
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath}).
		Publish(ctx, fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000)))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to: %v\n", ref)

	return nil
}
