package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/deploy"
	"github.com/spinnaker/spinctl/internal/halimport"
	"github.com/spinnaker/spinctl/internal/model"
	"github.com/spinnaker/spinctl/internal/tui"
)

var version = "dev"

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spinctl",
		Short: "Spinnaker configuration tool",
		Long:  "spinctl is a terminal UI tool for managing Spinnaker configuration and deployment.",
		RunE:  runTUI,
	}
	cmd.Version = version
	cmd.AddCommand(deployCmd())
	cmd.AddCommand(importCmd())
	cmd.AddCommand(showCmd())
	return cmd
}

func loadOrCreateConfig() (*config.SpinctlConfig, string) {
	configPath := config.DefaultConfigPath()
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		cfg = config.NewDefault()
	}
	return cfg, configPath
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, configPath := loadOrCreateConfig()
	lockPath := config.DefaultConfigDir() + "/.lock"
	lock, err := config.AcquireLock(lockPath)
	if err != nil {
		return err
	}
	defer config.ReleaseLock(lock)
	app := tui.NewApp(cfg, configPath)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func deployCmd() *cobra.Command {
	var services string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Spinnaker services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := loadOrCreateConfig()

			var filter []model.ServiceName
			if services != "" {
				for _, s := range strings.Split(services, ",") {
					name, err := model.ServiceNameFromString(strings.TrimSpace(s))
					if err != nil {
						return err
					}
					filter = append(filter, name)
				}
			}

			if errs := config.Validate(cfg); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "  validation: %s\n", e)
				}
				return fmt.Errorf("config validation failed")
			}

			cacheDir := filepath.Join(config.DefaultConfigDir(), "cache", "bom")
			fetcher := deploy.NewBOMFetcher(deploy.DefaultBOMURLPattern, cacheDir)
			bom, err := fetcher.Fetch(cfg.Version)
			if err != nil {
				return fmt.Errorf("fetching BOM: %w", err)
			}

			plan := deploy.BuildDeployPlan(filter)

			fmt.Println("Deploy plan:")
			for _, w := range plan.Warnings {
				fmt.Printf("  WARNING: %s\n", w)
			}
			for _, step := range plan.Steps {
				for _, svc := range step.Services {
					version, _ := bom.ServiceVersion(svc)
					fmt.Printf("  %s = %s\n", svc, version)
				}
			}

			if dryRun {
				return nil
			}

			fmt.Println("\nDeploying...")
			spinctlDir := config.DefaultConfigDir()
			stateFile := filepath.Join(spinctlDir, "deploy-state.json")

			if state, err := deploy.LoadDeployState(stateFile); err == nil {
				fmt.Printf("\nPrevious deploy interrupted. Completed: %v\n", state.Completed)
				deploy.RemoveDeployState(stateFile)
			}

			exec := &deploy.RealExecutor{}
			runner := deploy.NewDeployRunner(exec,
				"/opt/spinnaker/config",
				filepath.Join(spinctlDir, "deploy.log"),
				stateFile,
			)

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			results, err := runner.Run(ctx, cfg, bom, filter)
			for _, r := range results {
				status := "OK"
				if r.Err != nil {
					status = fmt.Sprintf("FAILED: %v", r.Err)
				}
				fmt.Printf("  %s %s (%s)\n", r.Service, status, r.Duration.Round(time.Millisecond))
			}
			return err
		},
	}

	cmd.Flags().StringVar(&services, "services", "", "Comma-separated list of services to deploy")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show deploy plan without executing")
	return cmd
}

func importCmd() *cobra.Command {
	var halDir string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Halyard configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if halDir == "" {
				detected := halimport.DetectHalDir()
				if detected == "" {
					return fmt.Errorf("no .hal directory found; use --hal-dir to specify")
				}
				halDir = detected
			}

			outputPath := config.DefaultConfigPath()
			result, err := halimport.Import(halimport.ImportOptions{
				HalDir:         halDir,
				DeploymentName: "default",
				OutputPath:     outputPath,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Import complete:\n")
			fmt.Printf("  Deployment: %s\n", result.DeploymentName)
			if result.BackupPath != "" {
				fmt.Printf("  Backup: %s\n", result.BackupPath)
			}
			if len(result.UnmappedFields) > 0 {
				fmt.Printf("  Unmapped fields: %s\n", strings.Join(result.UnmappedFields, ", "))
			}
			fmt.Printf("  Config saved to: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&halDir, "hal-dir", "", "Path to .hal directory (default: ~/.hal)")
	return cmd
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := config.DefaultConfigPath()
			_, err := config.LoadFromFile(configPath)
			if err != nil {
				return fmt.Errorf("no config found at %s: %w", configPath, err)
			}
			data, err := os.ReadFile(configPath)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}
}
