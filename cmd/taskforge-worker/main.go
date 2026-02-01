package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/config"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
	"github.com/shaeelbhatti2/taskforge/internal/worker"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "taskforge-worker"}
	root.AddCommand(startCmd())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func startCmd() *cobra.Command {
	var tagsRaw string
	var name string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a TaskForge worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			st, err := store.Open(cfg.DatabaseURL)
			if err != nil {
				return err
			}
			defer st.Close()
			if err := st.Migrate(context.Background()); err != nil {
				return err
			}
			nsID := cfg.DefaultNamespace()
			w := &domain.Worker{
				ID:          domain.NewID(),
				NamespaceID: nsID,
				Name:        name,
				Tags:        splitTags(tagsRaw),
				Capacity:    1,
			}
			reg := worker.NewRegistry(st)
			if err := reg.Register(context.Background(), w); err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			ticker := time.NewTicker(cfg.WorkerHeartbeat)
			defer ticker.Stop()
			slog.Info("worker started", "id", w.ID, "tags", w.Tags)
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					if err := reg.Heartbeat(ctx, nsID, w.ID); err != nil {
						slog.Error("heartbeat failed", "err", err)
					}
				}
			}
		},
	}
	cmd.Flags().StringVar(&tagsRaw, "tags", "", "comma-separated capacity tags")
	cmd.Flags().StringVar(&name, "name", "worker", "worker name")
	return cmd
}

func splitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
