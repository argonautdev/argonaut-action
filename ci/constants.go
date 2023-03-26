package main

import "os"

const (
	MIDGARD_URL  = "https://midgard.argonaut.dev"
	FRONTEGG_URL = "https://argonaut.frontegg.com"
)

type BuildType string

const (
	Docker    BuildType = "docker"
	BuildPack BuildType = "buildpack"
)

type ArtifactoryType string

const (
	CR ArtifactoryType = "cr" //container registry
)

type BuildRunStatus string

const (
	Requested BuildRunStatus = "requested"
	Triggered BuildRunStatus = "triggered"
	Running   BuildRunStatus = "running"
	Canceled  BuildRunStatus = "canceled"
	Failed    BuildRunStatus = "failed"
	Completed BuildRunStatus = "completed"
)

func GetMidgardUrl() string {
	host := os.Getenv("ARGONAUT_BACKEND")
	if host == "" {
		host = MIDGARD_URL
	}
	return host
}

func GetFrontEggUrl() string {
	host := os.Getenv("ARGONAUT_AUTH_SERVER")
	if host == "" {
		host = FRONTEGG_URL
	}
	return host
}
