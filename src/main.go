package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/webappsgo/casci/src/internal/config"
	"github.com/webappsgo/casci/src/pkg/cli"
	"github.com/webappsgo/casci/src/pkg/database"
	"github.com/webappsgo/casci/src/pkg/server"
)

// Build metadata injected at link time by the Makefile.
var (
	Version      = "dev"
	CommitID     = "unknown"
	BuildEpoch   = "0"
	OfficialSite = "https://casci.casapps.us"
)

// Exit codes follow the sysexits.h conventions the project standardises on.
const (
	exitOK          = 0
	exitUsage       = 64
	exitUnavailable = 69
	exitSoftware    = 70
	exitConfig      = 78
)

// shutdownTimeout bounds how long in-flight requests may finish draining.
const shutdownTimeout = 30 * time.Second

func main() {
	os.Exit(run())
}

// run holds the real entry point so that deferred cleanup still executes;
// os.Exit in main would skip it.
func run() int {
	// Publish the linker-stamped build metadata so every package reports the
	// same version without importing package main
	config.SetBuildInfo(Version, CommitID, BuildEpoch)

	cliCfg, err := cli.Parse(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cli.BinaryName(), err)
		return exitUsage
	}

	switch {
	case cliCfg.ShowHelp:
		cli.ShowHelp(Version)
		return exitOK
	case cliCfg.ShowVersion:
		cli.ShowVersion(Version, buildDate(), CommitID)
		return exitOK
	}

	if err := cliCfg.ApplyEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cli.BinaryName(), err)
		return exitConfig
	}

	if cliCfg.ShowStatus {
		return runStatus(cliCfg)
	}

	switch {
	case cliCfg.Shell != "":
		fmt.Fprintf(os.Stderr, "%s: shell integration is not implemented yet\n", cli.BinaryName())
		return exitUnavailable
	case cliCfg.Service != "":
		return reportError(cli.HandleService(cliCfg.Service))
	case cliCfg.Maintenance != "":
		return reportError(cli.HandleMaintenance(cliCfg.Maintenance, cliCfg.MaintenanceArg))
	case cliCfg.Update != "":
		return reportError(cli.HandleUpdate(cliCfg.Update, cliCfg.UpdateBranch))
	}

	return runServer(cliCfg)
}

// runStatus resolves the configured address and queries the health endpoint.
func runStatus(cliCfg *cli.Config) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cli.BinaryName(), err)
		return exitConfig
	}

	if err := cli.ShowStatus(cfg.Server.Host, cfg.Server.Port); err != nil {
		return 1
	}

	return exitOK
}

// runServer loads configuration, opens the database, and serves until a
// termination signal arrives.
func runServer(cliCfg *cli.Config) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cli.BinaryName(), err)
		return exitConfig
	}

	setupLogging(cliCfg.Debug || cfg.Server.Mode == "debug")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return exitUnavailable
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	srv, err := server.New(cfg, db)
	if err != nil {
		slog.Error("failed to initialize server", "error", err)
		return exitSoftware
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	slog.Info("server started",
		"version", Version,
		"commit", CommitID,
		"mode", cfg.Server.Mode,
		"address", cfg.Server.Host,
		"port", cfg.Server.Port,
		"site", OfficialSite,
	)

	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("server stopped", "error", err)
			return exitUnavailable
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		return exitSoftware
	}

	return exitOK
}

// setupLogging installs the structured JSON logger used across the project.
func setupLogging(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

// reportError prints a command error and maps it to an exit code.
func reportError(err error) int {
	if err == nil {
		return exitOK
	}

	fmt.Fprintf(os.Stderr, "%s: %v\n", cli.BinaryName(), err)
	return exitUnavailable
}

// buildDate renders the link-time epoch as an RFC 3339 UTC timestamp.
func buildDate() string {
	epoch, err := strconv.ParseInt(BuildEpoch, 10, 64)
	if err != nil || epoch == 0 {
		return "unknown"
	}
	return time.Unix(epoch, 0).UTC().Format(time.RFC3339)
}
