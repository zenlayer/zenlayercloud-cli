package loader

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenlayer/zenlayercloud-cli/internal/apiclient"
	"github.com/zenlayer/zenlayercloud-cli/internal/config"
	"github.com/zenlayer/zenlayercloud-cli/internal/output"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"gopkg.in/yaml.v3"
)

// apisFS holds the APIs filesystem. It is set by RegisterAll and may be
// overridden in tests via setTestFS.
var apisFS fs.FS

// productUsageTemplate is like Cobra's default but omits the " [command]" usage
// line for product commands, since we display "<api-name>" in Use instead.
const productUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Api List:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Api List:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} <api-name> --help" for more information about a command.{{end}}
`

// setTestFS replaces apisFS for use in unit tests.
func setTestFS(fsys fs.FS) { apisFS = fsys }

// RegisterAll scans the provided embedded APIs filesystem, parses each YAML
// definition, and registers cobra product sub-commands on root. Multiple APIs
// in the same product share one parent command.
//
// fsys must contain an "apis/en-US/" subtree (and optionally "apis/zh-CN/").
// Pass the embed.FS declared in the root main package.
func RegisterAll(root *cobra.Command, fsys embed.FS, getAccessKeyID, getAccessKeySecret func() string, getOutput, getQuery, getDebug, getEndpoint, getDryRun func() interface{}) error {
	apisFS = fsys
	lang := langDir(config.GetLanguage())

	// Collect en-US definitions to iterate over (source of truth for file list).
	enFiles, err := listYAMLFiles("apis/en-US")
	if err != nil {
		return fmt.Errorf("failed to list API definitions: %w", err)
	}

	// product name → product cobra command
	productCmds := make(map[string]*cobra.Command)

	for _, enPath := range enFiles {
		// enPath like "apis/en-US/zec/describe-instances.yaml"
		rel := strings.TrimPrefix(enPath, "apis/en-US/")
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) != 2 {
			continue
		}
		product := parts[0]
		filename := parts[1]

		def, err := loadWithFallback(product, filename, lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[WARNING] failed to load %s: %v\n", enPath, err)
			continue
		}

		// // Validate zh-CN consistency when language is zh-CN.
		// if lang == "zh-CN" {
		// 	enDef, _ := loadDefinition(fmt.Sprintf("apis/en-US/%s/%s", product, filename))
		// 	if enDef != nil {
		// 		_ = validateConsistency(enDef, def, fmt.Sprintf("zh-CN/%s/%s", product, filename))
		// 	}
		// }

		// Ensure product command exists.
		prodCmd, ok := productCmds[product]
		if !ok {
			prodCmd = &cobra.Command{
				Use:   product + " <api-name> [--param1 value1] [flags]",
				Short: fmt.Sprintf("Manage %s resources", strings.ToUpper(product)),
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) > 0 {
						return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
					}
					return cmd.Help()
				},
			}
			prodCmd.SetUsageTemplate(productUsageTemplate)
			// Use pager for product help output
			prodCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
				var buf strings.Builder
				c.SetOut(&buf)
				// Generate default help using usage template
				c.UsageFunc()(c)
				OutputWithPager(os.Stdout, buf.String())
			})
			productCmds[product] = prodCmd
			root.AddCommand(prodCmd)
		}

		apiCmd := makeAPICommand(def, getAccessKeyID, getAccessKeySecret, getOutput, getQuery, getDebug, getEndpoint, getDryRun)
		prodCmd.AddCommand(apiCmd)
	}

	return nil
}

// langDir maps config language codes to directory names.
func langDir(lang string) string {
	if lang == "zh" {
		return "zh-CN"
	}
	return "en-US"
}

// loadWithFallback loads the YAML for the given product/filename in the target
// language, falling back to en-US when the file doesn't exist.
func loadWithFallback(product, filename, lang string) (*APIDefinition, error) {
	path := fmt.Sprintf("apis/%s/%s/%s", lang, product, filename)
	if def, err := loadDefinition(path); err == nil {
		return def, nil
	}
	// Fallback to en-US.
	return loadDefinition(fmt.Sprintf("apis/en-US/%s/%s", product, filename))
}

// loadDefinition reads and parses one YAML file from the embedded FS.
func loadDefinition(path string) (*APIDefinition, error) {
	data, err := fs.ReadFile(apisFS, path)
	if err != nil {
		return nil, err
	}
	var def APIDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &def, nil
}

// listYAMLFiles returns all *.yaml paths under root within the embedded FS.
func listYAMLFiles(root string) ([]string, error) {
	var paths []string
	err := fs.WalkDir(apisFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

// makeAPICommand builds a cobra.Command for a single API definition.
func makeAPICommand(def *APIDefinition, getAccessKeyID, getAccessKeySecret func() string, getOutput, getQuery, getDebug, getEndpoint, getDryRun func() interface{}) *cobra.Command {
	// Build examples string for backward compatibility.
	var exampleLines []string
	for _, ex := range def.Examples {
		if ex.Desc != "" {
			exampleLines = append(exampleLines, "  # "+ex.Desc)
		}
		exampleLines = append(exampleLines, "  $ "+ex.Cmd)
	}

	cmd := &cobra.Command{
		Use:     def.Use,
		Short:   def.Short,
		Long:    def.Long,
		Example: strings.Join(exampleLines, "\n"),
	}

	// Use custom help function to generate structured help output with pager support
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		help := GenerateHelp(def, c.InheritedFlags())
		OutputWithPager(c.OutOrStdout(), help)
	})

	store := bindFlags(cmd, def)

	// Add --page-all only for commands that support pagination.
	var pageAll bool
	if isPaginatedDef(def) {
		cmd.Flags().BoolVar(&pageAll, "page-all", false, "Automatically fetch all pages and merge results into a single response")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Handle space-separated trailing args for array flags.
		if err := expandTrailingArgs(cmd, args, store); err != nil {
			return err
		}

		// Collect sensitive values from env (not from flags, to avoid exposure).
		sensitiveValues := collectSensitiveFromEnv(def)

		if err := validateRequired(store, def, sensitiveValues); err != nil {
			return err
		}

		params, err := collectParams(store, def, sensitiveValues)
		if err != nil {
			return err
		}

		// Dry-run: preview the request without sending it.
		if dryRun, ok := getDryRun().(bool); ok && dryRun {
			endpoint := "console.zenlayer.com"
			if ep, ok := getEndpoint().(string); ok && ep != "" {
				endpoint = ep
			}
			preview := map[string]interface{}{
				"dryRun":   true,
				"endpoint": endpoint,
				"service":  def.SDK.Service,
				"version":  def.SDK.Version,
				"action":   def.SDK.Action,
				"params":   params,
				"note":     "dry-run mode: request was not sent",
			}
			return output.FormatTo(os.Stdout, "json", preview)
		}

		keyID := getAccessKeyID()
		secret := getAccessKeySecret()

		cfg := common.NewConfig()
		if debug, ok := getDebug().(bool); ok && debug {
			t := true
			cfg.Debug = &t
		}
		if ep, ok := getEndpoint().(string); ok && ep != "" {
			cfg.Domain = ep
		}

		client, err := apiclient.NewCommonClient(keyID, secret, cfg)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		var result map[string]interface{}
		if pageAll {
			result, err = fetchAllPages(client, def, params)
		} else {
			result, err = client.Call(def.SDK.Service, def.SDK.Version, def.SDK.Action, params)
		}
		if err != nil {
			return err
		}

		outFmt := "json"
		if of, ok := getOutput().(string); ok && of != "" {
			outFmt = of
		}

		var toFormat interface{} = result
		if q, ok := getQuery().(string); ok && q != "" {
			filtered, err := output.ApplyQuery(q, result)
			if err != nil {
				return err
			}
			toFormat = filtered
		}

		return output.FormatTo(os.Stdout, outFmt, toFormat)
	}

	return cmd
}

// collectSensitiveFromEnv reads sensitive parameters from environment variables
// using the convention ZENLAYER_<UPPER_SNAKE_CASE_PARAM_NAME>.
func collectSensitiveFromEnv(def *APIDefinition) map[string]string {
	result := make(map[string]string)
	for i := range def.Parameters {
		param := &def.Parameters[i]
		if !param.Sensitive {
			continue
		}
		envKey := "ZENLAYER_" + strings.ToUpper(strings.ReplaceAll(param.Name, "-", "_"))
		if v := os.Getenv(envKey); v != "" {
			result[param.Name] = v
		}
	}
	return result
}
