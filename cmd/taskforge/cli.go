package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/config"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
	"github.com/shaeelbhatti2/taskforge/internal/workflow"
	"github.com/spf13/cobra"
)

func runCLI() error {
	root := &cobra.Command{Use: "taskforge"}
	root.AddCommand(jobCmd(), runCmd(), workflowCmd())
	return root.Execute()
}

func openStore() (store.Store, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	st, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	if err := st.Migrate(context.Background()); err != nil {
		st.Close()
		return nil, nil, err
	}
	return st, cfg, nil
}

func jobCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "job"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, cfg, err := openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			items, err := st.ListJobs(context.Background(), cfg.DefaultNamespace())
			if err != nil {
				return err
			}
			return printJSON(items)
		},
	}
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a shell job",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			command, _ := cmd.Flags().GetString("command")
			st, cfg, err := openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			now := time.Now().UTC()
			job := &domain.JobDefinition{
				ID: domain.NewID(), NamespaceID: cfg.DefaultNamespace(),
				Name: name, Type: domain.JobTypeShell, Command: command,
				TimeoutSec: 300, CreatedAt: now, UpdatedAt: now,
			}
			return st.CreateJob(context.Background(), job)
		},
	}
	create.Flags().String("name", "", "job name")
	create.Flags().String("command", "", "shell command")
	cmd.AddCommand(list, create)
	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "run"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, cfg, err := openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			items, err := st.ListJobRuns(context.Background(), store.RunFilter{
				NamespaceID: cfg.DefaultNamespace(), Limit: 50,
			})
			if err != nil {
				return err
			}
			return printJSON(items)
		},
	}
	logs := &cobra.Command{
		Use:   "logs",
		Short: "Show run logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("run id required")
			}
			st, cfg, err := openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			run, err := st.GetJobRun(context.Background(), cfg.DefaultNamespace(), args[0])
			if err != nil {
				return err
			}
			fmt.Println(run.Stdout)
			if run.Stderr != "" {
				fmt.Fprintln(os.Stderr, run.Stderr)
			}
			return nil
		},
	}
	cmd.AddCommand(list, logs)
	return cmd
}

func workflowCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "workflow"}
	validate := &cobra.Command{
		Use:   "validate",
		Short: "Validate workflow file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("file required")
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			var wf domain.Workflow
			if err := json.Unmarshal(data, &wf); err != nil {
				return err
			}
			return workflow.Validate(&wf)
		},
	}
	cmd.AddCommand(validate)
	return cmd
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
