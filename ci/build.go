package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"

	"dagger.io/dagger"
)

func build(context context.Context, buildRunId string, userRepoLoc string) error {

	fmt.Print("build task started!!")

	callbackPayload := &BuildRunCallbackPayload{
		ImageTag: "SHORT_SHA_TAG",
	}

	defer GetArgoClient().BuildRunCallback(buildRunId, callbackPayload)

	buildRunInfo, err := GetArgoClient().FetchBuildRunInfo(buildRunId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	buildInfo, err := GetArgoClient().FetchBuildInfo(buildRunInfo.BuildConfigId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	crAccess, err := GetArgoClient().FetchContainerRegistryAccess(buildInfo.ArtifactoryId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	fmt.Println("cr access response : ", *crAccess)

	execCmd := exec.CommandContext(context, "docker", "login", "--username", crAccess.Username, "--password", crAccess.Password, strings.TrimPrefix(crAccess.Url, "https://"))
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(out))
	}
	fmt.Printf(string(out))

	// initialize Dagger client
	client, err := dagger.Connect(context, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	defer client.Close()

	//cache := client.CacheVolume("argonaut")

	contextDir := client.Host().Directory(userRepoLoc)

	ref, err := client.Container().
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath}).
		Publish(context, fmt.Sprintf("%s/%s/%s:%.0f", strings.TrimPrefix(crAccess.Url, "https://"), "argonautdev", buildInfo.Name, math.Floor(rand.Float64()*10000000)))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	callbackPayload.Status = Completed

	fmt.Printf("build process over: %v\n", ref)

	return nil
}
