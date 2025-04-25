package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dek0valev/handyman/clients/openai"
	"github.com/dek0valev/handyman/clients/telegram"
	"github.com/dek0valev/handyman/consumer"
)

func main() {
	fmt.Println("Grab some coffee, make cool stuff!")

	telegramAuthToken := os.Getenv("TELEGRAM_AUTH_TOKEN")
	if telegramAuthToken == "" {
		log.Fatal("Please set the TELEGRAM_AUTH_TOKEN environment variable.")
	}

	telegramTargetChatIDStr := os.Getenv("TELEGRAM_TARGET_CHAT_ID")
	if telegramTargetChatIDStr == "" {
		log.Fatal("Please set the TELEGRAM_TARGET_CHAT_ID environment variable.")
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("Please set the OPENAI_API_KEY environment variable.")
	}

	telegramClient := telegram.NewClient(telegramAuthToken)
	openaiClient := openai.NewClient(openaiAPIKey)

	telegramTargetChatID, err := strconv.ParseInt(telegramTargetChatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_TARGET_CHAT_ID: %v", err)
	}

	c := consumer.NewConsumer(telegramClient, openaiClient, telegramTargetChatID)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go c.Run(ctx)

	<-sigChan

	cancel()

	fmt.Println("Goodbye!")
}
