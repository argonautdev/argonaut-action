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

	fmt.Println("fetch build run info complete : ", *buildRunInfo)

	buildInfo, err := GetArgoClient().FetchBuildInfo(buildRunInfo.BuildConfigId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	fmt.Println("fetch build info complete : ", *buildInfo)

	buildArgs, err := getBuildArgs(buildInfo.Id)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	crAccess, err := GetArgoClient().FetchContainerRegistryAccess(buildInfo.ArtifactoryId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	fmt.Println("cr access call success : ", crAccess.UrlWithPrefix)

	execCmd := exec.CommandContext(context, "docker", "login", "--username", crAccess.Username, "--password", crAccess.Password, strings.TrimPrefix(crAccess.Url, "https://"))
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(out))
	}

	fmt.Println("docker login complete : ", string(out))

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
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath, BuildArgs: buildArgs}).
		Publish(context, fmt.Sprintf("%s:%s", image, shortSha))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	callbackPayload.Status = Completed

	fmt.Printf("build process over: %v\n", ref)

	return nil
}

func getBuildArgs(buildConfigId string) ([]dagger.BuildArg, error) {
	res, err := GetArgoClient().FetchBuildTimeSecrets(buildConfigId)
	if err != nil {
		return nil, err
	}
	buildArgs := []dagger.BuildArg{}
	if res != nil {
		for _, secret := range res.BuildSecretsData.Data {
			buildArgs = append(buildArgs, dagger.BuildArg{
				Name:  secret.Key,
				Value: secret.Value,
			})
		}
	}
	return buildArgs, nil
}
