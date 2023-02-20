package main

const (
	MIDGARD_URL  = "https://c1b4-2405-201-6004-7855-99fc-c37d-9ef2-b902.in.ngrok.io"
	FRONTEGG_URL = "https://argonaut-pp.frontegg.com"
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
