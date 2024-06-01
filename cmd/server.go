package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/appclacks/server/config"
	"github.com/appclacks/server/internal/database"
	"github.com/appclacks/server/internal/http"
	"github.com/appclacks/server/internal/http/handlers"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func buildServerCmd(logger *slog.Logger) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Runs the HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			err := runServer(logger)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(2)
			}

		},
	}
	return serverCmd
}

func runServer(logger *slog.Logger) error {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("fail to read configuration file: %w", err)
	}
	var config config.Configuration
	if err := yaml.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("fail to parse yaml configuration file: %w", err)
	}
	store, err := database.New(logger, config.Database, config.Healthchecks.Probers)
	if err != nil {
		return err
	}
	healthcheckService := healthcheck.New(logger, store)
	handlersBuilder := handlers.NewBuilder(healthcheckService)
	server, err := http.NewServer(logger, config.HTTP, prometheus.DefaultRegisterer.(*prometheus.Registry), handlersBuilder)
	if err != nil {
		return err
	}
	signals := make(chan os.Signal, 1)
	errChan := make(chan error)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM)

	server.Start()
	go func() {
		for sig := range signals {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				logger.Info(fmt.Sprintf("received signal %s, starting shutdown", sig))
				signal.Stop(signals)
				err := server.Stop()
				if err != nil {
					errChan <- err
				}
				errChan <- nil
			}

		}
	}()
	exitErr := <-errChan
	return exitErr
}
