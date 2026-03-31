package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zenlayer/zenlayercloud-cli/internal/config"
	"github.com/zenlayer/zenlayercloud-cli/internal/output"
	"golang.org/x/term"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure zeno credentials and settings",
	Long: `Configure zeno credentials and settings interactively.

This command will prompt you for:
  - Profile name
  - Access Key ID
  - Access Key Secret
  - Language preference
  - Output format preference

The credentials are stored in ~/.zenlayer/credentials.json with restricted permissions.
Other settings are stored in ~/.zenlayer/config.json.`,
	RunE: runConfigureInteractive,
}

var configureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current profile configuration",
	Long:  `List all configuration settings for the current profile (excluding credentials).`,
	RunE:  runConfigureList,
}

var configureGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value by key.

Available keys:
  - profile: current profile name
  - language: language setting (en/zh)
  - output: output format (json/table)
  - access-key-id: access key ID
  - access-key-secret: access key secret`,
	Args:              cobra.ExactArgs(1),
	RunE:              runConfigureGet,
	ValidArgsFunction: completeConfigKeys,
}

var configureSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value by key.

Available keys:
  - profile: current profile name
  - language: language setting (en/zh)
  - output: output format (json/table)
  - access-key-id: access key ID
  - access-key-secret: access key secret`,
	Args:              cobra.ExactArgs(2),
	RunE:              runConfigureSet,
	ValidArgsFunction: completeConfigSet,
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.AddCommand(configureListCmd)
	configureCmd.AddCommand(configureGetCmd)
	configureCmd.AddCommand(configureSetCmd)
}

func runConfigureInteractive(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("Zenlayer Cloud CLI Configuration")
	fmt.Println()

	// Profile
	currentProfile := config.GetCurrentProfile()
	profile := promptWithDefault(reader, "Profile", currentProfile)

	// Ensure profile exists
	config.EnsureProfile(profile)
	config.SetCurrentProfile(profile)

	// Access Key ID
	currentKeyID := config.GetAccessKeyID()
	keyID := promptWithDefault(reader, "Access Key ID", currentKeyID)

	// Access Key Secret (hidden input)
	fmt.Print("Access Key Secret []: ")
	secretBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read secret: %w", err)
	}
	secret := strings.TrimSpace(string(secretBytes))
	if secret == "" {
		secret = config.GetAccessKeySecret()
	}

	// Language
	currentLang := config.GetLanguage()
	lang := promptWithDefault(reader, "Language (en/zh)", currentLang)
	if err := config.ValidateLanguage(lang); err != nil {
		return err
	}

	// Output format
	currentOutput := config.GetOutput()
	outputFmt := promptWithDefault(reader, "Output Format (json/table)", currentOutput)
	if err := config.ValidateOutput(outputFmt); err != nil {
		return err
	}

	// Save configuration
	config.SetProfileConfig(profile, lang, outputFmt)
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Save credentials
	config.SetCredentials(profile, keyID, secret)
	if err := config.SaveCredentials(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	credPath, _ := config.GetCredentialsPath()

	fmt.Println()
	fmt.Printf("Configuration saved to %s\n", configPath)
	fmt.Printf("Credentials saved to %s\n", credPath)

	return nil
}

func runConfigureList(cmd *cobra.Command, args []string) error {
	cfg := config.ListCurrentConfig()
	outputFormat := GetOutput()

	return output.Format(outputFormat, cfg)
}

func runConfigureGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := config.Get(key)

	if value == "" {
		return fmt.Errorf("configuration key '%s' is not set or unknown", key)
	}

	fmt.Println(value)
	return nil
}

func runConfigureSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate before setting
	switch key {
	case "language":
		if err := config.ValidateLanguage(value); err != nil {
			return err
		}
	case "output":
		if err := config.ValidateOutput(value); err != nil {
			return err
		}
	}

	if err := config.Set(key, value); err != nil {
		return err
	}

	fmt.Printf("Set '%s' to '%s'\n", key, value)
	return nil
}

var configKeys = []string{
	"profile\tcurrent profile name",
	"language\tlanguage setting (en/zh)",
	"output\toutput format (json/table)",
	"access-key-id\taccess key ID",
	"access-key-secret\taccess key secret",
}

func completeConfigKeys(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return configKeys, cobra.ShellCompDirectiveNoFileComp
}

func completeConfigSet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return configKeys, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return completeConfigValues(args[0]), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeConfigValues(key string) []string {
	switch key {
	case "language":
		return []string{"en\tEnglish", "zh\tChinese"}
	case "output":
		return []string{"json\tJSON format", "table\tTable format"}
	default:
		return nil
	}
}

func promptWithDefault(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s []: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}
