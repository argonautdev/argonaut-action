package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/pretty"
	"go.uber.org/zap"
)

type ArgoClient interface {
	FetchBuildRunInfo(buildRunId string) (*BuildRunInfo, error)
	FetchBuildInfo(buildId string) (*BuildConfigInfo, error)
	FetchContainerRegistryToken(crId string) (*RegistryToken, error)
}

type ArgoClientImpl struct {
	*resty.Client
	clientAuthInfo *GetClientIDAndSecretResponse
}

func (c *ArgoClientImpl) FetchBuildRunInfo(buildRunId string) (*BuildRunInfo, error) {
	out := BuildRunInfo{}
	resp, err := c.R().Get(fmt.Sprintf("/api/v1/build/run/%s", buildRunId))
	err = UnmarshalAndLog(resp, &out, err)
	return &out, err
}

func (c *ArgoClientImpl) FetchBuildInfo(buildId string) (*BuildConfigInfo, error) {
	// Hardcoding response for testing purposes
	out := BuildConfigInfo{
		Id:        buildId,
		Name:      "TEST",
		BuildType: Docker,
		Details: BuildConfigDetails{
			OCIBuildDetails: &OCIBuildDetails{
				DockerFilePath: "./Dockerfile",
			},
		},
	}
	//resp, err := c.R().Get(fmt.Sprintf("/api/v1/build/%s", buildId))
	//err = UnmarshalAndLog(resp, &out, err)
	return &out, nil
}

func (c *ArgoClientImpl) FetchContainerRegistryToken(crId string) (*RegistryToken, error) {
	panic("not implemented") // TODO: Implement
}

func GetMidgardClient(key, secret string) (ArgoClient, error) {

	argoClient := &ArgoClientImpl{Client: resty.New()}

	clientAuthInfo, err := getClientAuthInfo(key, secret)
	if err != nil {
		zap.S().Errorf("Could not construct client (internal err). Err: %v", err)
		return nil, err
	}

	argoClient.clientAuthInfo = clientAuthInfo

	argoClient.SetHeader("Authorization", clientAuthInfo.Accesstoken).SetBaseURL(MIDGARD_URL)

	return argoClient, nil

}

func getClientAuthInfo(key, secret string) (*GetClientIDAndSecretResponse, error) {
	resp, err := resty.New().SetBaseURL(FRONTEGG_URL).R().
		SetBody(&ApiTokenConfigStruct{
			ClientID:     key,
			ClientSecret: secret,
		}).
		Post("/identity/resources/auth/v1/api-token")
	if err != nil {
		zap.S().Errorf("Could not send request. Err: %v", err)
		return nil, err
	}
	if resp.IsError() {
		zap.S().Errorf("Could not send request. Err: %v", string(resp.Body()))
		return nil, errors.New("error status code")
	}

	var getClientIDAndSecretResponse GetClientIDAndSecretResponse
	err = json.Unmarshal(resp.Body(), &getClientIDAndSecretResponse)
	if err != nil {
		zap.S().Error("Could not convert reponse body. The following error occurred: %v", err)
		return nil, err
	}

	return &getClientIDAndSecretResponse, nil

}

func UnmarshalAndLog(resp *resty.Response, out interface{}, err error) error {
	err = LogResponseErrorOrRequestCreationError(resp, err)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp.Body(), out)
	if err != nil {
		zap.S().Errorf("Could not parse body, unexpected response type sent from server. Err: %v", err)
		return err
	}
	return nil

}

func LogResponseErrorOrRequestCreationError(resp *resty.Response, err error) error {
	if err != nil {
		zap.S().Errorf("Could not send request. Err: %v", err)
		return err
	}

	if resp.IsError() {
		zap.S().Errorf("Error status from server.\n%v", string(pretty.Color(pretty.Pretty(resp.Body()), nil)))
		return ErrCodeInResponse
	}

	return nil
}
