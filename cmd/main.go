package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"go-template/cmd/config"
	"go-template/internal/bot"
	"go-template/pkg/logger"
	psqlx "go-template/pkg/sqlx"
	ptelegram "go-template/pkg/telegram"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	log := logger.New(cfg.General.Debug)

	pgConnection, err := psqlx.NewDatabase(cfg.Postgres)
	if err != nil {
		log.Fatal("Failed to open postgres connection", zap.Error(err))
	}
	err = pgConnection.Ping()
	if err != nil {
		log.Fatal("Failed to ping postgres", zap.Error(err))
	}

	_ = psqlx.NewTransactor(pgConnection, log, &psqlx.TransactorConfig{Isolation: sql.LevelRepeatableRead})

	_ = psqlx.NewConnContainer(pgConnection, pgConnection)

	telegramClient, err := ptelegram.NewClient(log, cfg.Telegram)
	if err != nil {
		log.Fatal("Failed to init bot client", zap.Error(err))
	}

	telegramBot := bot.New(telegramClient)
	telegramBot.RegisterRoutes()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Bot client is starting...")
		telegramClient.Start()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	telegramClient.Stop()
	wg.Wait()

	log.Info("Bot was stopped gracefully")
	_ = log.Sync()
}
