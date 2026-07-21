package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-code-tracker-server/internal/httpapi"
	"ai-code-tracker-server/internal/migrate"
	"ai-code-tracker-server/internal/store"
)

type serverConfig struct {
	ListenAddress string
	MySQLDSN      string
}

func configFromEnv(getenv func(string) string) (serverConfig, error) {
	config := serverConfig{
		ListenAddress: getenv("LISTEN_ADDR"),
		MySQLDSN:      getenv("MYSQL_DSN"),
	}
	if config.ListenAddress == "" {
		config.ListenAddress = ":8080"
	}
	if config.MySQLDSN == "" {
		return serverConfig{}, errors.New("MYSQL_DSN is required")
	}
	return config, nil
}

func main() {
	config, err := configFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	database, err := sql.Open("mysql", config.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	startupContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(startupContext); err != nil {
		log.Fatal(fmt.Errorf("connect to MySQL: %w", err))
	}
	if err := migrate.Apply(startupContext, database); err != nil {
		log.Fatal(fmt.Errorf("apply migrations: %w", err))
	}

	server := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           httpapi.NewHandler(store.NewMySQLStore(database)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() { serverError <- server.ListenAndServe() }()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-shutdownContext.Done():
		context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(context); err != nil {
			log.Printf("shutdown server: %v", err)
		}
	}
}
