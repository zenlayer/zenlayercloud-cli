package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var completionUninstall bool

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate or uninstall shell completion script",
	Long: `Generate shell completion script for the specified shell.

Bash:
  # Linux
  zeno completion bash > /etc/bash_completion.d/zeno
  # macOS (requires bash-completion)
  zeno completion bash > $(brew --prefix)/etc/bash_completion.d/zeno

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  echo "autoload -U compinit; compinit" >> ~/.zshrc

  zeno completion zsh > "${fpath[1]}/_zeno"

  # You will need to start a new shell for this setup to take effect.

Fish:
  zeno completion fish > ~/.config/fish/completions/zeno.fish

PowerShell:
  zeno completion powershell | Out-String | Invoke-Expression

Uninstall:
  zeno completion --uninstall              # uninstall all
  zeno completion --uninstall [bash|zsh|fish|powershell]  # uninstall specific shell`,
	DisableFlagsInUseLine: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		// Include all flags (local + inherited) so --uninstall gets tab completion
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			comps = append(comps, "--"+f.Name+"\t"+f.Usage)
		})
		// Include shell names (bash, zsh, fish, powershell)
		comps = append(comps, getCompletionShellNames()...)
		return comps, cobra.ShellCompDirectiveNoFileComp
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if completionUninstall {
			if len(args) == 0 {
				return nil
			}
			if len(args) > 1 {
				return fmt.Errorf("too many arguments")
			}
			if _, ok := getCompletionInstaller(args[0]); ok {
				return nil
			}
			return fmt.Errorf("invalid shell %q, must be one of: bash, zsh, fish, powershell", args[0])
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if completionUninstall {
			return runCompletionUninstall(args)
		}
		inst, ok := getCompletionInstaller(args[0])
		if !ok {
			return nil
		}
		return inst.Install(cmd.Root(), os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.Flags().BoolVar(&completionUninstall, "uninstall", false, "uninstall shell completion (all shells if no shell specified)")
}

func runCompletionUninstall(args []string) error {
	var installers []completionInstaller
	if len(args) == 1 {
		inst, ok := getCompletionInstaller(args[0])
		if !ok {
			return nil
		}
		installers = []completionInstaller{inst}
	} else {
		installers = getAllCompletionInstallers()
	}

	var anyRemoved bool
	for _, inst := range installers {
		removed, err := inst.Uninstall()
		if err != nil {
			return err
		}
		if len(removed) == 0 {
			fmt.Fprintf(os.Stderr, "No zeno completion found for %s in standard locations.\n", inst.GetName())
			continue
		}
		anyRemoved = true
		for _, p := range removed {
			fmt.Printf("Removed: %s\n", p)
		}
	}
	if anyRemoved {
		fmt.Println("Restart your shell for changes to take effect.")
	}
	return nil
}
