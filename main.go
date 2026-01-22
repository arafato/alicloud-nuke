package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/arafato/ali-nuke/config"
	"github.com/arafato/ali-nuke/infrastructure"
	_ "github.com/arafato/ali-nuke/resources"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
	"github.com/arafato/ali-nuke/version"
	"github.com/spf13/cobra"
)

// Global flags that can be used across commands
var (
	configFile      string
	accessKeyID     string
	accessKeySecret string
	noDryRun        bool
	shortVersion    bool
)

var rootCmd = &cobra.Command{
	Use:   "ali-nuke",
	Short: "ali-nuke removes every resource from your Alibaba Cloud account",
	Long:  `A tool which removes every resource from an Alibaba Cloud account. Use it with caution, since it cannot distinguish between production and non-production.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ali-nuke",
	Run: func(cmd *cobra.Command, args []string) {
		if shortVersion {
			version.PrintShort(os.Stdout)
		} else {
			version.Print(os.Stdout)
		}
	},
}

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Execute nuke operation",
	Long: `Nuke command performs destructive operations based on the provided configuration.
Use with caution and review the dry-run output before executing.`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if accessKeyID == "" {
			return fmt.Errorf("--access-key-id is required")
		}
		if accessKeySecret == "" {
			return fmt.Errorf("--access-key-secret is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		executeNuke()
	},
}

func init() {
	rootCmd.AddCommand(nukeCmd)
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print short version string")

	nukeCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (required)")
	nukeCmd.Flags().StringVar(&accessKeyID, "access-key-id", "", "Alibaba Cloud Access Key ID (required)")
	nukeCmd.Flags().StringVar(&accessKeySecret, "access-key-secret", "", "Alibaba Cloud Access Key Secret (required)")
	nukeCmd.Flags().BoolVar(&noDryRun, "no-dry-run", false, "Execute without dry run (actually delete resources)")

	nukeCmd.MarkFlagRequired("config")
	nukeCmd.MarkFlagRequired("access-key-id")
	nukeCmd.MarkFlagRequired("access-key-secret")
}

func executeNuke() {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	creds := &types.Credentials{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	}

	// Dynamically fetch all regions and apply exclusions
	fmt.Println("Fetching available regions...")
	regions, err := utils.GetActiveRegions(creds, cfg.Regions.Excludes)
	if err != nil {
		log.Fatalf("Error fetching regions: %v", err)
	}
	fmt.Printf("Scanning %d regions (excluded %d)...\n", len(regions), len(cfg.Regions.Excludes))

	resources := infrastructure.ProcessCollection(creds, regions)
	infrastructure.FilterCollection(resources, cfg)

	visibleCount := resources.VisibleCount()
	fmt.Printf("Scan complete: Found %d resources in total. To be removed %d, Filtered %d\n",
		visibleCount, resources.NumOf(types.Ready), resources.NumOf(types.Filtered))
	utils.PrettyPrintStatus(resources)

	if !noDryRun {
		fmt.Println("Dry run complete.")
		return
	}

	fmt.Println("Executing actual nuke operation... do you really want to continue (yes/no)?")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Nuke operation aborted.")
		return
	}
	fmt.Println("Nuke operation confirmed.")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start printer goroutine BEFORE removal to show progress during the operation
	wg.Add(1)
	go utils.PrintStatusWithContext(&wg, ctx, resources)

	if err := infrastructure.RemoveCollection(ctx, resources); err != nil {
		log.Printf("Error removing resources: %v", err)
	}

	// Cancel printer after removal completes, then wait for it to finish
	cancel()
	wg.Wait()

	// Print final summary
	failedCount := resources.NumOf(types.Failed)
	deletedCount := resources.NumOf(types.Deleted)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Process finished. Deleted: %d, Failed: %d\n", deletedCount, failedCount)

	if failedCount > 0 {
		fmt.Println("\nFailed resources:")
		for _, resource := range resources {
			if resource.State() == types.Failed {
				fmt.Printf("  - [%s] %s: %s (%s)\n", resource.Region, resource.ProductName, resource.ResourceName, resource.ResourceID)
			}
		}
		fmt.Println("\nNote: Some resources may have failed due to dependencies. Run again to retry.")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
