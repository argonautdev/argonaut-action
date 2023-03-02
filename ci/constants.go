package main

const (
	MIDGARD_URL  = "https://midgard-1.pp.argonaut.live"
	FRONTEGG_URL = "https://argonaut1-pp.us.frontegg.com"
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
