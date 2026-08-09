package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/netview/netview/internal/api"
	"github.com/netview/netview/internal/config"
	"github.com/netview/netview/internal/db"
	"github.com/netview/netview/internal/storage"
)

// version 由构建时通过 -ldflags "-X main.version=..." 注入。
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "打印版本并退出")
	flag.Parse()
	if *showVersion {
		log.Printf("NetView %s", version)
		return
	}

	cfg := config.Load()

	database, err := db.Connect(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Printf("NetView %s: database migrations applied", version)

	store, err := storage.New(cfg.Storage.DataDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	server := api.NewServer(cfg, database, store)
	router := server.Router()

	addr := cfg.Server.Host + ":" + itoa(cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("NetView %s listening on http://%s", version, addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

