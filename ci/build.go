package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"
)

func build(context context.Context, buildRunId string, userRepoLoc string) error {

	fmt.Println("build task started!!")

	callbackPayload := &BuildRunCallbackPayload{
		Status: Running,
	}

	defer GetArgoClient().BuildRunCallback(buildRunId, callbackPayload)

	shortSha := os.Getenv("SHORT_SHA")

	if shortSha == "" {
		callbackPayload.Status = Failed
		return errors.New("image tag not generated")
	}

	callbackPayload.ImageTag = fmt.Sprintf("%s-%s", shortSha, time.Now().Format("01020304"))

	fmt.Printf("short sha : [%s]", shortSha)

	buildRunInfo, err := GetArgoClient().FetchBuildRunInfo(buildRunId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	fmt.Printf("fetch build run info complete : [%v] \n", *buildRunInfo)

	buildInfo, err := GetArgoClient().FetchBuildInfo(buildRunInfo.BuildConfigId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	fmt.Printf("fetch build info complete : [%v] \n", *buildInfo)

	buildArgs, err := getBuildArgs(buildInfo.Id)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	fmt.Printf("fetch build args complete : Count[%d] \n", len(buildArgs))

	crAccess, err := GetArgoClient().FetchContainerRegistryAccess(buildInfo.ArtifactoryId)
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}
	fmt.Printf("cr access call success : [%s]  \n", crAccess.UrlWithPrefix)

	execCmd := exec.CommandContext(context, "docker", "login", "--username", crAccess.Username, "--password", crAccess.Password, strings.TrimPrefix(crAccess.Url, "https://"))
	out, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("docker login failed : [%s]  \n", string(out))
		return fmt.Errorf(string(out))
	}

	fmt.Printf("docker login complete : [%s] \n", string(out))

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

	workingDir := filepath.Join(userRepoLoc, buildInfo.Details.OCIBuildDetails.WorkingDir)

	contextDir := client.Host().Directory(workingDir)

	ref, err := client.Container().
		Build(contextDir, dagger.ContainerBuildOpts{Dockerfile: buildInfo.Details.OCIBuildDetails.DockerFilePath, BuildArgs: buildArgs}).
		Publish(context, fmt.Sprintf("%s:%s", image, callbackPayload.ImageTag))
	if err != nil {
		callbackPayload.Status = Failed
		return err
	}

	callbackPayload.Status = Completed

	fmt.Printf("build process over: %s \n", ref)

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
