package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/pretty"
	"go.uber.org/zap"
)

type ArgoClient interface {
	FetchBuildRunInfo(buildRunId string) (*BuildRun, error)
	FetchBuildInfo(buildId string) (*BuildConfig, error)
	FetchContainerRegistryAccess(crId string) (*RegistryAccess, error)
	FetchBuildTimeSecrets(buildConfigId string) (*BuildSecretFetch, error)
	BuildRunCallback(buildRunId string, payload *BuildRunCallbackPayload) error
}

type ArgoClientImpl struct {
	*resty.Client
	clientAuthInfo *GetClientIDAndSecretResponse
}

func (c *ArgoClientImpl) FetchBuildRunInfo(buildRunId string) (*BuildRun, error) {
	out := BuildRun{}
	resp, err := c.R().Get(fmt.Sprintf("/api/v1/build/run/%s", buildRunId))
	err = UnmarshalAndLog(resp, &out, err)
	return &out, err
}

func (c *ArgoClientImpl) FetchBuildInfo(buildId string) (*BuildConfig, error) {
	out := BuildConfig{}
	resp, err := c.R().Get(fmt.Sprintf("/api/v1/build/%s", buildId))
	err = UnmarshalAndLog(resp, &out, err)
	return &out, err
}
func (c *ArgoClientImpl) BuildRunCallback(buildRunId string, payload *BuildRunCallbackPayload) error {
	resp, err := c.R().SetBody(*payload).Post(fmt.Sprintf("/api/v1/build/run/%s/callback", buildRunId))
	err = UnmarshalAndLog(resp, map[string]interface{}{}, err)
	return err
}

func (c *ArgoClientImpl) FetchContainerRegistryAccess(crId string) (*RegistryAccess, error) {
	out := RegistryAccess{}
	resp, err := c.R().Get(fmt.Sprintf("/api/v1/registries/%s/access", crId))
	err = UnmarshalAndLog(resp, &out, err)
	return &out, err
}

func (c *ArgoClientImpl) FetchBuildTimeSecrets(buildConfigId string) (*BuildSecretFetch, error) {
	out := BuildSecretFetch{}
	resp, err := c.R().Get(fmt.Sprintf("/api/v1/build/%s/secrets", buildConfigId))
	err = UnmarshalAndLog(resp, &out, err)
	return &out, err
}

var argoClientInstance ArgoClient = nil

func GetArgoClient() ArgoClient {
	if argoClientInstance == nil {
		panic("server client not initialized yet")
	}
	return argoClientInstance
}

func InitializeArgoClient() (ArgoClient, error) {

	argoClient := &ArgoClientImpl{Client: resty.New()}

	switch {
	default:
		key := os.Getenv("ARG_AUTH_KEY")
		secret := os.Getenv("ARG_AUTH_SECRET")
		if key == "" || secret == "" {
			return nil, errors.New("access to argonaut server is not configured")
		}
		clientAuthInfo, err := getFEAuthInfo(key, secret)
		if err != nil {
			zap.S().Errorf("Could not construct client (internal err). Err: %v", err)
			return nil, err
		}
		argoClient.clientAuthInfo = clientAuthInfo
		argoClient.SetHeader("Authorization", clientAuthInfo.Accesstoken)
	}

	argoClient.SetBaseURL(MIDGARD_URL)

	argoClient.SetRetryCount(2).
		AddRetryCondition(func(res *resty.Response, reqErr error) bool {
			zap.S().Info("Trace Info : ")
			if res != nil && res.Request != nil {
				ti := res.Request.TraceInfo()
				zap.S().Infof("  Content size  : %v", res.Request.Header["Content-Length"])
				zap.S().Infof("  DNSLookup     : %v", ti.DNSLookup)
				zap.S().Infof("  ConnTime      : %v", ti.ConnTime)
				zap.S().Infof("  TCPConnTime   : %v", ti.TCPConnTime)
				zap.S().Infof("  TLSHandshake  : %v", ti.TLSHandshake)
				zap.S().Infof("  ServerTime    : %v", ti.ServerTime)
				zap.S().Infof("  ResponseTime  : %v", ti.ResponseTime)
				zap.S().Infof("  TotalTime     : %v", ti.TotalTime)
				zap.S().Infof("  IsConnReused  : %v", ti.IsConnReused)
				zap.S().Infof("  IsConnWasIdle : %v", ti.IsConnWasIdle)
				zap.S().Infof("  ConnIdleTime  : %v", ti.ConnIdleTime)

				zap.S().Infof("  Resp Time       :", res.Time())
				zap.S().Infof("  Resp Received At:", res.ReceivedAt())
			}

			if reqErr != nil {
				zap.S().Errorf("Request trace info for err. Err: %v", reqErr)
				return true
				// if errors.Is(reqErr, syscall.ECONNRESET) {
				// 	zap.S().Error("  Retrying Request!")
				// 	return true
				// }
			}
			return false
		},
		).EnableTrace().SetContentLength(true).SetRetryWaitTime(1000)

	argoClientInstance = argoClient

	return argoClient, nil

}

func getFEAuthInfo(key, secret string) (*GetClientIDAndSecretResponse, error) {
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
		return nil, errors.New("authentication error : " + string(resp.Body()))
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
