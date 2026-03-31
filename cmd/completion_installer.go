package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// completionInstaller defines the interface for shell-specific completion install/uninstall.
type completionInstaller interface {
	GetName() string
	Install(root *cobra.Command, w io.Writer) error
	Uninstall() (removed []string, err error)
}

var completionInstallers = map[string]completionInstaller{
	"bash":       &bashInstaller{},
	"zsh":        &zshInstaller{},
	"fish":       &fishInstaller{},
	"powershell": &powershellInstaller{},
}

func getCompletionInstaller(name string) (completionInstaller, bool) {
	inst, ok := completionInstallers[name]
	return inst, ok
}

func getCompletionShellNames() []string {
	return []string{"bash", "zsh", "fish", "powershell"}
}

func getAllCompletionInstallers() []completionInstaller {
	names := []string{"bash", "zsh", "fish", "powershell"}
	result := make([]completionInstaller, 0, len(names))
	for _, n := range names {
		if inst, ok := completionInstallers[n]; ok {
			result = append(result, inst)
		}
	}
	return result
}

// bashInstaller handles bash completion.
type bashInstaller struct{}

func (b *bashInstaller) GetName() string { return "bash" }

func (b *bashInstaller) Install(root *cobra.Command, w io.Writer) error {
	return root.GenBashCompletionV2(w, true)
}

func (b *bashInstaller) Uninstall() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	paths := []string{
		"/etc/bash_completion.d/zeno",
		"/usr/local/etc/bash_completion.d/zeno",
		"/opt/homebrew/etc/bash_completion.d/zeno",
		filepath.Join(home, ".bash_completion.d", "zeno"),
	}
	return removePaths(paths), nil
}

// zshInstaller handles zsh completion.
type zshInstaller struct{}

func (z *zshInstaller) GetName() string { return "zsh" }

func (z *zshInstaller) Install(root *cobra.Command, w io.Writer) error {
	return root.GenZshCompletion(w)
}

func (z *zshInstaller) Uninstall() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	paths := []string{
		filepath.Join(home, ".zsh", "completions", "_zeno"),
		filepath.Join(home, ".oh-my-zsh", "completions", "_zeno"),
		"/usr/local/share/zsh/site-functions/_zeno",
		"/usr/share/zsh/site-functions/_zeno",
		"/opt/homebrew/share/zsh/site-functions/_zeno",
	}
	// Also search all directories currently in fpath via the FPATH env var
	if fpath := os.Getenv("FPATH"); fpath != "" {
		for _, dir := range strings.Split(fpath, ":") {
			paths = append(paths, filepath.Join(dir, "_zeno"))
		}
	}
	return removePaths(paths), nil
}

// fishInstaller handles fish completion.
type fishInstaller struct{}

func (f *fishInstaller) GetName() string { return "fish" }

func (f *fishInstaller) Install(root *cobra.Command, w io.Writer) error {
	return root.GenFishCompletion(w, true)
}

func (f *fishInstaller) Uninstall() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	p := filepath.Join(home, ".config", "fish", "completions", "zeno.fish")
	return removePaths([]string{p}), nil
}

// powershellInstaller handles PowerShell completion.
type powershellInstaller struct{}

func (p *powershellInstaller) GetName() string { return "powershell" }

func (p *powershellInstaller) Install(root *cobra.Command, w io.Writer) error {
	return root.GenPowerShellCompletionWithDesc(w)
}

func (p *powershellInstaller) Uninstall() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	profilePath := os.Getenv("USERPROFILE")
	if profilePath == "" {
		profilePath = home
	}
	profilePaths := []string{
		filepath.Join(profilePath, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
		filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"),
		filepath.Join(home, ".config", "powershell", "profile.ps1"),
	}
	for _, path := range profilePaths {
		if _, err := os.Stat(path); err == nil {
			return removePowerShellProfile(path)
		}
	}
	return nil, nil
}

// removePaths removes files at the given paths. Returns list of successfully removed paths.
func removePaths(paths []string) []string {
	var removed []string
	for _, p := range paths {
		if err := os.Remove(p); err == nil {
			removed = append(removed, p)
		}
	}
	return removed
}

func removePowerShellProfile(profilePath string) ([]string, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, "zeno") && !strings.Contains(line, "zenlayercloud-cli") {
			newLines = append(newLines, line)
		}
	}
	if len(newLines) == len(lines) {
		return nil, nil
	}
	if err := os.WriteFile(profilePath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("failed to write profile: %w", err)
	}
	return []string{profilePath + " (zeno lines removed)"}, nil
}
