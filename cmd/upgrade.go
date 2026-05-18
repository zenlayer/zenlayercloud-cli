package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenlayer/zenlayercloud-cli/internal/updater"
	"github.com/zenlayer/zenlayercloud-cli/internal/version"
)

var (
	upgradeCheck    bool
	upgradeList     bool
	upgradeVersion  string
	upgradeRollback bool
	upgradeYes      bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade zeno to the latest version",
	Long: `Upgrade the zeno CLI to the latest version from GitHub Releases.

The current binary is backed up as <path>.bak before replacement,
enabling rollback with --rollback if needed.`,
	RunE: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&upgradeCheck, "check", false, "check for a new version without upgrading")
	upgradeCmd.Flags().BoolVar(&upgradeList, "list", false, "list all available versions")
	upgradeCmd.Flags().StringVar(&upgradeVersion, "version", "", "upgrade to a specific version (e.g. v1.0.8)")
	upgradeCmd.Flags().BoolVar(&upgradeRollback, "rollback", false, "roll back to the previous backup version")
	upgradeCmd.Flags().BoolVarP(&upgradeYes, "yes", "y", false, "skip confirmation prompt")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	if upgradeRollback {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine binary path: %w", err)
		}
		fmt.Println("Rolling back to previous version...")
		if err := updater.Rollback(exe); err != nil {
			return err
		}
		fmt.Println("Rollback successful.")
		return nil
	}

	u := updater.New()

	if upgradeList {
		releases, err := u.FetchAll()
		if err != nil {
			return err
		}
		current := strings.TrimPrefix(version.Version, "v")
		fmt.Printf("%-16s %-3s %s\n", "Version", "", "Published")
		fmt.Printf("%-16s %-3s %s\n", strings.Repeat("-", 16), "", strings.Repeat("-", 10))
		for _, r := range releases {
			marker := ""
			if strings.TrimPrefix(r.TagName, "v") == current {
				marker = "*"
			}
			fmt.Printf("%-16s %-3s %s\n", r.TagName, strings.TrimSpace(marker), r.PublishedAt.Format("2006-01-02"))
		}
		return nil
	}

	// Determine target version tag.
	targetTag := upgradeVersion
	if targetTag != "" {
		if !strings.HasPrefix(targetTag, "v") {
			targetTag = "v" + targetTag
		}
	} else {
		latest, err := u.FetchLatest()
		if err != nil {
			return err
		}
		targetTag = latest.TagName
	}

	current := version.Version
	cur := strings.TrimPrefix(current, "v")
	target := strings.TrimPrefix(targetTag, "v")

	if upgradeCheck {
		fmt.Printf("Current version: %s\n", current)
		label := "Latest version: "
		if upgradeVersion != "" {
			label = "Target version: "
		}
		fmt.Printf("%s%s\n", label, targetTag)
		if updater.CompareVersions(cur, target) >= 0 {
			fmt.Println("Already up to date.")
		} else {
			fmt.Printf("Update available: %s → %s\n", current, targetTag)
		}
		return nil
	}

	if updater.CompareVersions(cur, target) >= 0 {
		fmt.Printf("Already up to date (%s).\n", current)
		return nil
	}

	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Target version:  %s\n", targetTag)

	if !upgradeYes {
		fmt.Print("Proceed with upgrade? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	return performUpgrade(targetTag)
}

func performUpgrade(tag string) error {
	archive := updater.ArchiveName(tag)
	dlURL := updater.DownloadURL(tag, archive)
	csURL := updater.ChecksumURL(tag)

	fmt.Printf("Downloading %s...\n", archive)
	archivePath, err := updater.Download(dlURL)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	fmt.Println("Verifying checksum...")
	if err := updater.VerifyChecksum(archivePath, csURL, archive); err != nil {
		return err
	}
	fmt.Println("Checksum verified.")

	fmt.Println("Extracting...")
	tmpDir, err := os.MkdirTemp("", "zeno-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	newBin, err := updater.ExtractBinary(archivePath, tmpDir)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current binary path: %w", err)
	}

	fmt.Printf("Installing to %s...\n", exe)
	if err := updater.Install(newBin, exe); err != nil {
		return err
	}

	fmt.Printf("Successfully upgraded zeno to %s\n", tag)
	return nil
}
