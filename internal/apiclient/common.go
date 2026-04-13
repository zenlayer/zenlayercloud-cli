// Package apiclient provides a generic API client that wraps the Zenlayer SDK
// common.Client, enabling dynamic API calls via parameter maps without requiring
// generated code for each endpoint.
package apiclient

import (
	"encoding/json"
	"fmt"

	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

// CommonClient wraps the SDK common.Client, inheriting authentication, signing,
// retry, and debug capabilities without modification.
type CommonClient struct {
	common.Client
}

// NewCommonClient creates a CommonClient with the given credentials and config.
// If cfg is nil, SDK defaults are used.
func NewCommonClient(keyID, secret string, cfg *common.Config) (*CommonClient, error) {
	if keyID == "" || secret == "" {
		return nil, fmt.Errorf("access key ID and secret are required")
	}
	c := &CommonClient{}
	if err := c.InitWithCredential(common.NewCredential(keyID, secret)); err != nil {
		return nil, fmt.Errorf("failed to initialize credentials: %w", err)
	}
	if cfg == nil {
		cfg = common.NewConfig()
	}
	if err := c.WithConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply config: %w", err)
	}
	c.WithRequestClient("zeno")
	return c, nil
}

// commonRequest implements the common.Request interface using an embedded
// BaseRequest for all transport/signing methods, with a payload map as the
// JSON body so no per-API request struct is needed.
type commonRequest struct {
	*common.BaseRequest
	payload map[string]interface{}
}

func newCommonRequest(service, version, action string, payload map[string]interface{}) *commonRequest {
	req := &commonRequest{
		BaseRequest: &common.BaseRequest{},
		payload:     payload,
	}
	req.Init().InitWithApiInfo(service, version, action)
	return req
}

// MarshalJSON overrides default struct marshalling so ApiCall serialises only
// the caller's payload map as the HTTP request body.
func (r *commonRequest) MarshalJSON() ([]byte, error) {
	if r.payload == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(r.payload)
}

// commonResponse implements common.Response by embedding BaseResponse (for error
// parsing) while capturing the successful API response into a generic map.
type commonResponse struct {
	*common.BaseResponse
	RequestId string                 `json:"requestId"`
	Response  map[string]interface{} `json:"response"`
}

// Call executes a Zenlayer API endpoint identified by service/version/action,
// passing params as the JSON body. Returns the response.Response map on success.
func (c *CommonClient) Call(service, version, action string, params map[string]interface{}) (map[string]interface{}, error) {
	req := newCommonRequest(service, version, action, params)
	resp := &commonResponse{BaseResponse: &common.BaseResponse{}}
	if err := c.ApiCall(req, resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}
