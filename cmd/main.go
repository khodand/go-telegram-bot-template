package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"go-telegram-bot-template/cmd/config"
	"go-telegram-bot-template/internal/bot"
	"go-telegram-bot-template/pkg/logger"
	ptelegram "go-telegram-bot-template/pkg/telegram"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	log := logger.New(cfg.General.Debug)

	telegramClient, err := ptelegram.NewClient(log, cfg.Telegram)
	if err != nil {
		log.Fatal("failed to init bot client", zap.Error(err))
	}

	telegramBot := bot.New(telegramClient)
	telegramBot.RegisterRoutes()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("bot client is starting...")
		telegramClient.Start()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	telegramClient.Stop()
	wg.Wait()

	log.Info("bot was stopped gracefully")
	_ = log.Sync()
}
