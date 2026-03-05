package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shaeelbhatti2/taskforge/internal/api"
	"github.com/shaeelbhatti2/taskforge/internal/auth"
	"github.com/shaeelbhatti2/taskforge/internal/config"
	"github.com/shaeelbhatti2/taskforge/internal/logging"
	"github.com/shaeelbhatti2/taskforge/internal/metrics"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] != "server" {
		if err := runCLI(); err != nil {
			slog.Error("cli failed", "err", err)
			os.Exit(1)
		}
		return
	}
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}
	logging.Setup(cfg.LogLevel)
	st, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("store open failed", "err", err)
		os.Exit(1)
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.Migrate(ctx); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}
	authSvc := auth.New(cfg.APIKey())
	srv := api.NewServer(st, authSvc)
	mux := http.NewServeMux()
	mux.Handle("/", srv.Handler())
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("web"))))
	mux.HandleFunc("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	})
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: mux}
	go func() {
		slog.Info("taskforge listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()
	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()
	_ = server.Shutdown(context.Background())
}
