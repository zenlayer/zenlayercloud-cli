package apiclient

import (
	"encoding/json"
	"testing"

	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func TestNewCommonClient_MissingCredentials(t *testing.T) {
	_, err := NewCommonClient("", "", nil)
	if err == nil {
		t.Error("NewCommonClient() with empty credentials should return error")
	}
}

func TestNewCommonClient_ValidCredentials(t *testing.T) {
	// Uses fake credentials — no network call is made here.
	client, err := NewCommonClient("test-key-id", "test-secret", nil)
	if err != nil {
		t.Fatalf("NewCommonClient() error = %v", err)
	}
	if client == nil {
		t.Error("NewCommonClient() returned nil client")
	}
}

func TestNewCommonClient_WithConfig(t *testing.T) {
	cfg := common.NewConfig()
	client, err := NewCommonClient("key", "secret", cfg)
	if err != nil {
		t.Fatalf("NewCommonClient() with config error = %v", err)
	}
	if client == nil {
		t.Error("NewCommonClient() with config returned nil client")
	}
}

func TestCommonRequest_MarshalJSON_NilPayload(t *testing.T) {
	req := newCommonRequest("zec", "2024-04-01", "TestAction", nil)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("MarshalJSON(nil payload) = %q, want '{}'", string(data))
	}
}

func TestCommonRequest_MarshalJSON_WithPayload(t *testing.T) {
	payload := map[string]interface{}{
		"zoneId":       "asia-east-1a",
		"instanceType": "z2a.cpu.1",
	}
	req := newCommonRequest("zec", "2024-04-01", "CreateZecInstances", payload)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal MarshalJSON result error = %v", err)
	}
	if result["zoneId"] != "asia-east-1a" {
		t.Errorf("zoneId = %v, want 'asia-east-1a'", result["zoneId"])
	}
}

func TestCommonRequest_InitWithApiInfo(t *testing.T) {
	req := newCommonRequest("bmc", "2022-11-20", "DescribeInstances", nil)
	if req.GetAction() != "DescribeInstances" {
		t.Errorf("GetAction() = %q, want 'DescribeInstances'", req.GetAction())
	}
	if req.GetApiVersion() != "2022-11-20" {
		t.Errorf("GetApiVersion() = %q, want '2022-11-20'", req.GetApiVersion())
	}
	if req.GetServiceName() != "bmc" {
		t.Errorf("GetServiceName() = %q, want 'bmc'", req.GetServiceName())
	}
}
