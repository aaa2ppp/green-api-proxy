package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"green-api-proxy/internal/config"
	"green-api-proxy/internal/logger"
	"green-api-proxy/internal/proxy"
	"green-api-proxy/internal/server"
	"green-api-proxy/internal/www"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	logger.SetupDefault(cfg.Logger)
	log.Printf("logging level is %v", cfg.Logger.Level)

	proxy, err := proxy.New(cfg.Proxy)
	if err != nil {
		log.Fatalf("create proxy failed: %v", err)
	}

	server, err := server.New(cfg.Server, proxy, www.New())
	if err != nil {
		log.Fatalf("create server failed: %v", err)
	}

	// Запускаем сервер
	go func() {
		var err error
		if cfg.Server.TLS {
			log.Printf("server listens on %s (HTTPS)", server.Addr)
			err = server.ListenAndServeTLS("./cert/cert.pem", "./cert/key.pem")
		} else {
			log.Printf("server listens on %s (HTTP)", server.Addr)
			err = server.ListenAndServe()
		}
		if err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// gracefully shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received.
	sig := <-ch
	log.Printf("shutdown by signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("gracefully shutdown failed: %v", err)
	}
}
