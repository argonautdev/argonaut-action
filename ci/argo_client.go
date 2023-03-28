package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/pretty"
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
			fmt.Printf("Could not construct client (internal err). Err: %v \n", err)
			return nil, err
		}
		argoClient.clientAuthInfo = clientAuthInfo
		argoClient.SetHeader("Authorization", clientAuthInfo.Accesstoken)
	}

	argoClient.SetBaseURL(GetMidgardUrl())

	argoClient.SetRetryCount(2).
		AddRetryCondition(func(res *resty.Response, reqErr error) bool {

			if reqErr != nil {
				fmt.Printf("Request trace info for err. Err: %v  \n", reqErr)
				fmt.Printf("Trace Info :  \n")
				if res != nil && res.Request != nil {
					ti := res.Request.TraceInfo()
					fmt.Printf("  Content size  : %v  \n", res.Request.Header["Content-Length"])
					fmt.Printf("  DNSLookup     : %v  \n", ti.DNSLookup)
					fmt.Printf("  ConnTime      : %v  \n", ti.ConnTime)
					fmt.Printf("  TCPConnTime   : %v  \n", ti.TCPConnTime)
					fmt.Printf("  TLSHandshake  : %v  \n", ti.TLSHandshake)
					fmt.Printf("  ServerTime    : %v  \n", ti.ServerTime)
					fmt.Printf("  ResponseTime  : %v \n", ti.ResponseTime)
					fmt.Printf("  TotalTime     : %v \n", ti.TotalTime)
					fmt.Printf("  IsConnReused  : %v \n", ti.IsConnReused)
					fmt.Printf("  IsConnWasIdle : %v \n", ti.IsConnWasIdle)
					fmt.Printf("  ConnIdleTime  : %v \n", ti.ConnIdleTime)
					fmt.Printf("  Resp Time       : %v \n", res.Time())
					fmt.Printf("  Resp Received At: %v \n", res.ReceivedAt())
				}
				return true
			}
			return false
		},
		).EnableTrace().SetContentLength(true).SetRetryWaitTime(1000)

	argoClientInstance = argoClient

	return argoClient, nil

}

func getFEAuthInfo(key, secret string) (*GetClientIDAndSecretResponse, error) {
	resp, err := resty.New().SetBaseURL(GetFrontEggUrl()).R().
		SetBody(&ApiTokenConfigStruct{
			ClientID:     key,
			ClientSecret: secret,
		}).
		Post("/identity/resources/auth/v1/api-token")
	if err != nil {
		fmt.Printf("Could not send request. Err: %v  \n", err)
		return nil, err
	}
	if resp.IsError() {
		fmt.Printf("Could not send request. Err: %v  \n", string(resp.Body()))
		return nil, errors.New("authentication error : " + string(resp.Body()))
	}

	var getClientIDAndSecretResponse GetClientIDAndSecretResponse
	err = json.Unmarshal(resp.Body(), &getClientIDAndSecretResponse)
	if err != nil {
		fmt.Printf("Could not convert reponse body. The following error occurred: %v \n", err)
		return nil, err
	}

	fmt.Print("Authentication successful! \n")
	return &getClientIDAndSecretResponse, nil

}

func UnmarshalAndLog(resp *resty.Response, out interface{}, err error) error {
	err = LogResponseErrorOrRequestCreationError(resp, err)
	if err != nil {
		return err
	}
	body := resp.Body()
	if len(body) > 0 {
		err = json.Unmarshal(body, out)
		if err != nil {
			fmt.Printf("Could not parse body, unexpected response type sent from server. Err: %v  \n", err)
			return err
		}
	}
	return nil

}

func LogResponseErrorOrRequestCreationError(resp *resty.Response, err error) error {
	if err != nil {
		fmt.Printf("Could not send request. Err: %v  \n", err)
		return err
	}

	if resp.IsError() {
		fmt.Printf("Error status from server.\n%v  \n", string(pretty.Color(pretty.Pretty(resp.Body()), nil)))
		return ErrCodeInResponse
	}

	return nil
}
