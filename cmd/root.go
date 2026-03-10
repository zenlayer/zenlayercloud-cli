// Package cmd contains all CLI commands for zencli.
package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zenlayer/zenlayercloud-cli/internal/config"
	"github.com/zenlayer/zenlayercloud-cli/internal/loader"
)

var (
	cfgProfile      string
	cfgOutput       string
	cfgAccessKeyID  string
	cfgAccessSecret string
	cfgEndpoint     string
	cfgDebug        bool
	apisFS          embed.FS
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "zencli",
	Short: "Zenlayer Cloud Command Line Interface",
	Long: `zencli is the command line interface for Zenlayer Cloud API.

It provides commands to manage Zenlayer Cloud resources including
bare metal servers, virtual machines, networking, and more.

To get started, run 'zencli configure' to set up your credentials.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// fsys must be the embed.FS containing the apis/ directory, declared in main.
func Execute(fsys embed.FS) error {
	apisFS = fsys
	registerProductCommands()
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags (including -h/--help so it appears under Global Flags for all subcommands)
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for this command")
	rootCmd.PersistentFlags().StringVarP(&cfgProfile, "profile", "p", "", "profile name to use (default: from config)")
	rootCmd.PersistentFlags().StringVarP(&cfgOutput, "output", "o", "", "output format: json, table (default: from config)")
	rootCmd.PersistentFlags().StringVar(&cfgAccessKeyID, "access-key-id", "", "access key ID (overrides config)")
	rootCmd.PersistentFlags().StringVar(&cfgAccessSecret, "access-key-secret", "", "access key secret (overrides config)")
	rootCmd.PersistentFlags().StringVar(&cfgEndpoint, "endpoint", "", "API domain/endpoint (overrides default)")
	rootCmd.PersistentFlags().BoolVar(&cfgDebug, "debug", false, "enable debug mode")

	// Flag value completions
	rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json\tJSON format", "table\tTable format"}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.GetAllProfiles(), cobra.ShellCompDirectiveNoFileComp
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Check for debug mode from environment variable
	if os.Getenv("ZENLAYER_DEBUG") == "true" {
		cfgDebug = true
	}

	// Load configuration
	if err := config.Load(); err != nil {
		if cfgDebug {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		}
	}

	// Apply command line overrides
	if cfgProfile != "" {
		config.SetCurrentProfile(cfgProfile)
	}

	// Override from environment variables
	if envProfile := os.Getenv("ZENLAYER_PROFILE"); envProfile != "" && cfgProfile == "" {
		config.SetCurrentProfile(envProfile)
	}
}

// GetDebug returns whether debug mode is enabled.
func GetDebug() bool {
	return cfgDebug
}

// GetOutput returns the output format, applying priority: CLI flag > env > config.
func GetOutput() string {
	if cfgOutput != "" {
		return cfgOutput
	}
	return config.GetOutput()
}

// GetAccessKeyID returns the access key ID, applying priority: CLI flag > env > config.
func GetAccessKeyID() string {
	if cfgAccessKeyID != "" {
		return cfgAccessKeyID
	}
	if envKey := os.Getenv("ZENLAYER_ACCESS_KEY_ID"); envKey != "" {
		return envKey
	}
	return config.GetAccessKeyID()
}

// registerProductCommands loads all YAML-defined product API commands into rootCmd.
func registerProductCommands() {
	err := loader.RegisterAll(
		rootCmd,
		apisFS,
		GetAccessKeyID,
		GetAccessKeySecret,
		func() interface{} { return GetOutput() },
		func() interface{} { return GetDebug() },
		func() interface{} { return GetEndpoint() },
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register product commands: %v\n", err)
	}
}

// GetEndpoint returns the API domain/endpoint, applying priority: CLI flag > env.
func GetEndpoint() string {
	if cfgEndpoint != "" {
		return cfgEndpoint
	}
	if envEndpoint := os.Getenv("ZENLAYER_ENDPOINT"); envEndpoint != "" {
		return envEndpoint
	}
	return ""
}

// GetAccessKeySecret returns the access key secret, applying priority: CLI flag > env > config.
func GetAccessKeySecret() string {
	if cfgAccessSecret != "" {
		return cfgAccessSecret
	}
	if envSecret := os.Getenv("ZENLAYER_ACCESS_KEY_SECRET"); envSecret != "" {
		return envSecret
	}
	return config.GetAccessKeySecret()
}
