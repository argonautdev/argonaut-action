package main

import "time"

// *************** Auth *********************

type ApiTokenConfigStruct struct {
	ClientID     string `yaml:"key" json:"clientId"`
	ClientSecret string `yaml:"secret" json:"secret"`
}

type GetClientIDAndSecretResponse struct {
	Expires      string `json:"expires"`
	Expiresin    int    `json:"expiresIn"`
	Accesstoken  string `json:"accessToken"`
	Refreshtoken string `json:"refreshToken"`
}

// *************** Build **********************

type BuildConfig struct {
	Id              string             `json:"id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	Name            string             `json:"name"`
	Disable         bool               `json:"disable"`
	OrganizationId  string             `json:"organization_id"`
	CIIntegrationId string             `json:"ci_integration_id"`
	RepoId          string             `json:"repo_id"`
	BuildType       BuildType          `json:"build_type" validate:"required" enums:"docker"`
	Details         BuildConfigDetails `json:"details"`
	ArtifactoryType ArtifactoryType    `json:"artifactory_type" validate:"required" enums:"cr"`
	ArtifactoryId   string             `json:"artifactory_id"`
}

type BuildConfigDetails struct {
	OCIBuildDetails OCIBuildDetails `json:"oci_build_details" validate:"required"`
}

type OCIBuildDetails struct {
	DockerFilePath string `json:"docker_file_path" validate:"required"`
	WorkingDir     string `json:"working_dir"`
}

type BuildRun struct {
	Id              string          `json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	BuildConfigId   string          `json:"build_config_id"`
	Status          BuildRunStatus  `json:"status" validate:"required" enums:"requested,triggered,running,canceled,failed,completed"`
	CIRef           string          `json:"ci_ref"`
	ArtifactoryType ArtifactoryType `json:"artifactory_type" validate:"required" enums:"cr"`
	ArtifactoryId   string          `json:"artifactory_id"`
	RepoMeta        RepoMeta        `json:"repo_meta"`
	BinaryOutput    BinaryOutput    `json:"binary_output"`
	TriggeredBy     string          `json:"triggered_by"`
	OrganizationId  string          `json:"organization_id"`
	PipelineRunId   string          `json:"pipeline_run_id"`
}

type BinaryOutput struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type RepoMeta struct {
	Branch    string `json:"branch" validate:"required"`
	CommitSha string `json:"commit_sha"`
	Message   string `json:"message"`
	Username  string `json:"username"`
}

type BuildRunCallbackPayload struct {
	Image    string         `json:"image"`
	ImageTag string         `json:"image_tag"`
	Status   BuildRunStatus `json:"status"`
	Error    string         `json:"error"`
}

// ************* Container Registry ************

type RegistryAccess struct {
	Username      string     `json:"username"`
	Password      string     `json:"password"`
	Url           string     `json:"url"`
	ExpiresAt     *time.Time `json:"expires_at"`
	UrlWithPrefix string     `json:"url_with_prefix"`
}

//************* Secrets ********************
type BuildSecretFetch struct {
	SecretId         string           `json:"secret_id"`
	BuildSecretsData BuildSecretsData `json:"build_secrets_data"`
}

type BuildSecretsData struct {
	Data []BuildSecret `json:"data"`
}

type BuildSecret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
