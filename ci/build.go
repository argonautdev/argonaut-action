package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"

	"dagger.io/dagger"
)

func build(context context.Context, buildRunId string, userRepoLoc string) error {
	buildRunInfo, err := GetArgoClient().FetchBuildRunInfo(buildRunId)
	if err != nil {
		return err
	}

	buildInfo, err := GetArgoClient().FetchBuildInfo(buildRunInfo.BuildConfigId)
	if err != nil {
		return err
	}

	// initialize Dagger client
	client, err := dagger.Connect(context, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	//cache := client.CacheVolume("argonaut")

	contextDir := client.Host().Directory(userRepoLoc)

	status := Completed

	ref, err := client.Container().
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath}).
		Publish(context, fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000)))
	if err != nil {
		status = Failed
	}

	GetArgoClient().BuildRunCallback(buildRunId, BuildRunCallbackPayload{
		ImageTag: "SHORT_SHA_TAG",
		Status:   status,
	})

	fmt.Printf("build process over: %v %v\n", ref, status)

	return err
}
