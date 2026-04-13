package loader

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// minimalYAML is a minimal valid API definition for use in tests.
const minimalYAML = `
name: DescribeZecInstances
product: zec
use: describe-instances
short: "Query ZEC instances"
long: "Query ZEC instances."
sdk:
  service: zec
  version: 2024-04-01
  action: DescribeZecInstances
parameters:
  - name: zone-id
    type: string
    description: "Availability zone ID"
`

const minimalYAMLZH = `
name: DescribeZecInstances
product: zec
use: describe-instances
short: "查询 ZEC 实例"
long: "查询 ZEC 实例。"
sdk:
  service: zec
  version: 2024-04-01
  action: DescribeZecInstances
parameters:
  - name: zone-id
    type: string
    description: "可用区 ID"
`

// testFS returns a MapFS that mimics the apis/ directory structure.
func testFS() fstest.MapFS {
	return fstest.MapFS{
		"apis/en-US/zec/describe-instances.yaml": &fstest.MapFile{Data: []byte(minimalYAML)},
		"apis/zh-CN/zec/describe-instances.yaml": &fstest.MapFile{Data: []byte(minimalYAMLZH)},
		"apis/en-US/bmc/describe-instances.yaml": &fstest.MapFile{Data: []byte(`
name: DescribeInstances
product: bmc
use: describe-instances
short: "Query BMC instances"
long: "Query BMC instances."
sdk:
  service: bmc
  version: 2022-11-20
  action: DescribeInstances
parameters: []
`)},
	}
}

func init() {
	// Inject test filesystem so loader functions work without embed.
	setTestFS(testFS())
}

func TestLangDir(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"zh", "zh-CN"},
		{"en", "en-US"},
		{"", "en-US"},
		{"fr", "en-US"},
	}
	for _, tt := range tests {
		got := langDir(tt.input)
		if got != tt.want {
			t.Errorf("langDir(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListYAMLFiles(t *testing.T) {
	files, err := listYAMLFiles("apis/en-US")
	if err != nil {
		t.Fatalf("listYAMLFiles() error = %v", err)
	}
	if len(files) == 0 {
		t.Fatal("listYAMLFiles() returned no files")
	}
	for _, f := range files {
		if len(f) == 0 {
			t.Error("empty file path returned")
		}
	}
}

func TestLoadDefinition(t *testing.T) {
	t.Run("load existing en-US file", func(t *testing.T) {
		def, err := loadDefinition("apis/en-US/zec/describe-instances.yaml")
		if err != nil {
			t.Fatalf("loadDefinition() error = %v", err)
		}
		if def.Name == "" {
			t.Error("expected non-empty Name")
		}
		if def.SDK.Service != "zec" {
			t.Errorf("SDK.Service = %q, want 'zec'", def.SDK.Service)
		}
	})

	t.Run("load non-existent file returns error", func(t *testing.T) {
		_, err := loadDefinition("apis/en-US/zec/nonexistent.yaml")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestLoadWithFallback(t *testing.T) {
	t.Run("zh-CN file exists and loads", func(t *testing.T) {
		def, err := loadWithFallback("zec", "describe-instances.yaml", "zh-CN")
		if err != nil {
			t.Fatalf("loadWithFallback() error = %v", err)
		}
		if def.Name == "" {
			t.Error("expected non-empty Name")
		}
	})

	t.Run("missing zh-CN file falls back to en-US", func(t *testing.T) {
		// bmc only has an en-US file in the test FS.
		def, err := loadWithFallback("bmc", "describe-instances.yaml", "zh-CN")
		if err != nil {
			t.Fatalf("loadWithFallback() error = %v", err)
		}
		if def.Name == "" {
			t.Error("expected non-empty Name from fallback")
		}
	})
}

func TestMakeAPICommand_Basic(t *testing.T) {
	def := &APIDefinition{
		Use:   "describe-instances",
		Short: "Query instances",
		Long:  "Query instances in detail.",
		SDK:   SDKInfo{Service: "zec", Version: "2024-04-01", Action: "DescribeZecInstances"},
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Description: "Zone ID"},
			{Name: "page-size", Type: "integer", Description: "Page size"},
		},
	}

	cmd := makeAPICommand(def,
		func() string { return "" },
		func() string { return "" },
		func() interface{} { return "" },
		func() interface{} { return "" },    // getQuery
		func() interface{} { return false },
		func() interface{} { return "" },
		func() interface{} { return false }, // getDryRun
	)

	if cmd.Use != "describe-instances" {
		t.Errorf("Use = %q, want 'describe-instances'", cmd.Use)
	}
	if cmd.Short != "Query instances" {
		t.Errorf("Short = %q, want 'Query instances'", cmd.Short)
	}
	if cmd.Long != "Query instances in detail." {
		t.Errorf("Long = %q, want 'Query instances in detail.'", cmd.Long)
	}
	if cmd.Flags().Lookup("zone-id") == nil {
		t.Error("expected --zone-id flag to be registered")
	}
	if cmd.Flags().Lookup("page-size") == nil {
		t.Error("expected --page-size flag to be registered")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

func TestMakeAPICommand_WithExamples(t *testing.T) {
	def := &APIDefinition{
		Use:   "create-instances",
		Short: "Create instances",
		Examples: []Example{
			{Desc: "Create a basic instance", Cmd: "zlcloud zec create-instances --zone-id cn-east-1"},
			{Cmd: "zlcloud zec create-instances --zone-id cn-west-1"},
		},
	}

	cmd := makeAPICommand(def,
		func() string { return "" },
		func() string { return "" },
		func() interface{} { return "" },
		func() interface{} { return "" },    // getQuery
		func() interface{} { return false },
		func() interface{} { return "" },
		func() interface{} { return false }, // getDryRun
	)

	if !strings.Contains(cmd.Example, "Create a basic instance") {
		t.Errorf("Example should contain description, got: %q", cmd.Example)
	}
	if !strings.Contains(cmd.Example, "zlcloud zec create-instances") {
		t.Errorf("Example should contain command, got: %q", cmd.Example)
	}
	// Second example has no Desc, so only the command line should appear.
	if !strings.Contains(cmd.Example, "--zone-id cn-west-1") {
		t.Errorf("Example should contain second command, got: %q", cmd.Example)
	}
}

func TestMakeAPICommand_NoExamples(t *testing.T) {
	def := &APIDefinition{
		Use:      "list-items",
		Short:    "List items",
		Examples: nil,
	}

	cmd := makeAPICommand(def,
		func() string { return "" },
		func() string { return "" },
		func() interface{} { return "" },
		func() interface{} { return "" },    // getQuery
		func() interface{} { return false },
		func() interface{} { return "" },
		func() interface{} { return false }, // getDryRun
	)

	if cmd.Example != "" {
		t.Errorf("expected empty Example, got %q", cmd.Example)
	}
}

func TestMakeAPICommand_DryRun(t *testing.T) {
	def := &APIDefinition{
		Use:   "describe-instances",
		Short: "Query instances",
		SDK:   SDKInfo{Service: "zec", Version: "2024-04-01", Action: "DescribeZecInstances"},
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Description: "Zone ID"},
		},
	}

	cmd := makeAPICommand(def,
		func() string { return "" },
		func() string { return "" },
		func() interface{} { return "json" },
		func() interface{} { return "" },    // getQuery
		func() interface{} { return false },
		func() interface{} { return "" },
		func() interface{} { return true }, // getDryRun = true
	)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{"--zone-id", "SEL-A"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, `"dryRun"`) {
		t.Errorf("expected dryRun field in output, got: %s", out)
	}
	if !strings.Contains(out, `"DescribeZecInstances"`) {
		t.Errorf("expected action in output, got: %s", out)
	}
	if !strings.Contains(out, `"zoneId"`) {
		t.Errorf("expected param zoneId in output, got: %s", out)
	}
	if !strings.Contains(out, `"SEL-A"`) {
		t.Errorf("expected param value SEL-A in output, got: %s", out)
	}
}

func TestCollectSensitiveFromEnv(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "password", Type: "string", Sensitive: true},
			{Name: "zone-id", Type: "string", Sensitive: false},
		},
	}

	t.Setenv("ZENLAYER_PASSWORD", "secret123")

	result := collectSensitiveFromEnv(def)
	if result["password"] != "secret123" {
		t.Errorf("password = %q, want 'secret123'", result["password"])
	}
	if _, ok := result["zone-id"]; ok {
		t.Error("zone-id should not appear in sensitive values")
	}
}
