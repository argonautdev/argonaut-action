package main

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

type BuildConfigInfo struct {
	Id        string             `json:"id"`
	Name      string             `json:"name"`
	RepoId    string             `json:"repo_id"`
	BuildType BuildType          `json:"build_type"`
	Details   BuildConfigDetails `json:"details"`
}

type BuildConfigDetails struct {
	OCIBuildDetails *OCIBuildDetails
}

type OCIBuildDetails struct {
	DockerFilePath string
}

type BuildRunInfo struct {
	Id              string
	BuildConfigId   string
	Status          BuildRunStatus
	CIRef           string
	ArtifactoryType ArtifactoryType
	ArtifactoryId   string
	RepoMeta        RepoMeta
	BinaryOutput    BinaryOutput
	OrganizationId  string
}

type BinaryOutput struct {
	Name string
	Tag  string
}

type RepoMeta struct {
	Branch    string
	CommitSha string
	Message   string
	Username  string
}

// ************* Container Registry ************

type RegistryToken struct {
}
