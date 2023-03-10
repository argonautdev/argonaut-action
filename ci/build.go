package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dagger.io/dagger"
)

func build(context context.Context, buildRunId string, userRepoLoc string) error {

	fmt.Print("build task started!!")

	shortSha := os.Getenv("SHORT_SHA")

	if shortSha == "" {
		return errors.New("image tag not generated")
	}

	callbackPayload := &BuildRunCallbackPayload{
		ImageTag: shortSha,
	}

	defer GetArgoClient().BuildRunCallback(buildRunId, callbackPayload)

	buildRunInfo, err := GetArgoClient().FetchBuildRunInfo(buildRunId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	fmt.Println("build run info call success : ", *buildRunInfo)

	buildInfo, err := GetArgoClient().FetchBuildInfo(buildRunInfo.BuildConfigId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	fmt.Println("build info call success : ", *buildInfo)

	crAccess, err := GetArgoClient().FetchContainerRegistryAccess(buildInfo.ArtifactoryId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	fmt.Println("cr access call success : ", *crAccess)

	execCmd := exec.CommandContext(context, "docker", "login", "--username", crAccess.Username, "--password", crAccess.Password, strings.TrimPrefix(crAccess.Url, "https://"))
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(out))
	}

	fmt.Println("docker login success : ", string(out))

	// initialize Dagger client
	client, err := dagger.Connect(context, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	defer client.Close()

	//cache := client.CacheVolume("argonaut")

	image := fmt.Sprintf("%s/%s", strings.TrimPrefix(crAccess.UrlWithPrefix, "https://"), buildInfo.Name)
	callbackPayload.Image = image

	contextDir := client.Host().Directory(userRepoLoc)

	ref, err := client.Container().
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath}).
		Publish(context, fmt.Sprintf("%s:%s", image, shortSha))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	callbackPayload.Status = Completed

	fmt.Printf("build process over: %v\n", ref)

	return nil
}
