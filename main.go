package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"

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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all supported resource types",
	Long:  "List all resources types that are currently supported by this version of ali-nuke.",
	Run: func(cmd *cobra.Command, args []string) {

		collectors := infrastructure.ListCollectors()
		for _, name := range collectors {
			fmt.Println(name)
		}

		fmt.Println("Total:", len(collectors))
	},
}

func init() {
	rootCmd.AddCommand(nukeCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)

	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print short version string")

	nukeCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file. If not provided no exclude filters are set.")
	nukeCmd.Flags().StringVar(&accessKeyID, "access-key-id", "", "Alibaba Cloud Access Key ID (required)")
	nukeCmd.Flags().StringVar(&accessKeySecret, "access-key-secret", "", "Alibaba Cloud Access Key Secret (required)")
	nukeCmd.Flags().BoolVar(&noDryRun, "no-dry-run", false, "Execute without dry run (actually delete resources)")

	nukeCmd.MarkFlagRequired("access-key-id")
	nukeCmd.MarkFlagRequired("access-key-secret")
}

func executeNuke() {
	var cfg *config.Config
	if configFile == "" {
		c := config.NewConfig()
		cfg = &c
	} else {
		var err error
		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
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

	// Initialize logger for collecting warnings/errors
	logger := utils.NewScanLogger()

	// Start spinner animation
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Scanning %d regions (excluded %d)...", len(regions), len(cfg.Regions.Excludes))
	s.Start()

	scanStart := time.Now()
	resources := infrastructure.ProcessCollection(creds, regions, logger)
	infrastructure.FilterCollection(resources, cfg)
	scanDuration := time.Since(scanStart)

	// Stop spinner before printing results
	s.Stop()

	visibleCount := resources.VisibleCount()
	fmt.Printf("Scan complete in %s: Found %d resources in total. To be removed %d, Filtered %d\n",
		formatDuration(scanDuration), visibleCount, resources.NumOf(types.Ready), resources.NumOf(types.Filtered))
	utils.PrettyPrintStatus(resources)

	// Flush logs to file and print summary if there were warnings/errors
	if logger.HasEntries() {
		if err := logger.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write log file: %v\n", err)
		}
		logger.PrintSummary()
	}

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

// formatDuration formats a duration in a human-readable way.
// For durations < 60s, it shows seconds (e.g., "45s").
// For durations >= 60s, it shows minutes and seconds (e.g., "1m42s").
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
