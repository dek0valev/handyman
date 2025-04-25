package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dek0valev/handyman/clients/openai"
	"github.com/dek0valev/handyman/clients/telegram"
	"github.com/dek0valev/handyman/entities"
)

type Consumer struct {
	telegramClient *telegram.Client
	openaiClient   *openai.Client
	targetChatID   int64
}

func NewConsumer(telegramClient *telegram.Client, openaiClient *openai.Client, targetChatID int64) *Consumer {
	return &Consumer{
		telegramClient: telegramClient,
		openaiClient:   openaiClient,
		targetChatID:   targetChatID,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	updatesChan := make(chan telegram.Update, 100)
	offset := int64(0)

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updates, err := c.fetchUpdates(offset)
				if err != nil {
					continue
				}

				for _, update := range updates {
					updatesChan <- update
					offset = update.ID + 1
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updatesChan:
			if err := c.processUpdate(update); err != nil {
				log.Printf("Error processing update: %v", err)
				continue
			}
		}
	}
}

func (c *Consumer) fetchUpdates(offset int64) ([]telegram.Update, error) {
	updates, err := c.telegramClient.GetUpdates(offset, 100)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func (c *Consumer) processUpdate(update telegram.Update) error {
	if update.Message == nil {
		return nil
	}

	if update.Message.Chat.ID != c.targetChatID {
		return nil
	}

	if update.Message.Audio != nil {
		return c.handleHoroscopeAudio(update)
	}

	return nil
}

func (c *Consumer) handleHoroscopeAudio(update telegram.Update) error {
	const prompt = `Твоя задача – выделить присланный текст в JSON формат.

	Не добавляй ничего лишнего, никакого Markdown, никаких пояснений.
	Четко выделяй то, что содержится в аудио-транскрипции.

	Каждый элемент JSON должен содержать:
	- "zodiac_sign" - знак зодиака
	- "content" - гороскоп для знака зодиака

	Итоговый JSON должен содержать массив объектов, каждый из которых соответствует одному знаку зодиака.

	Выводи только те знаки зодиака, которые есть в списке ниже:
	- "Овны"
	- "Девы"
	- "Скорпионы"
	- "Козероги"
	- "Рыбы"
	`

	if !strings.Contains(update.Message.Audio.FileName, "Гороскоп") {
		return nil
	}

	audioFileInfo, err := c.telegramClient.GetFile(update.Message.Audio.FileID)
	if err != nil {
		return err
	}

	audioFileBytes, err := c.telegramClient.DownloadFile(audioFileInfo.FilePath)
	if err != nil {
		return err
	}

	transcription, err := c.openaiClient.TranscribeAudio(audioFileBytes, update.Message.Audio.FileName)
	if err != nil {
		return err
	}

	horoscopesResponse, err := c.openaiClient.GenerateResponse(prompt, transcription)
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	var horoscopes []entities.Horoscope
	if err := json.Unmarshal([]byte(horoscopesResponse), &horoscopes); err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	loc := time.FixedZone("UTC+07:00", 7*60*60)
	today := time.Now().In(loc)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Гороскоп на <b>%s</b>\n\n", today.Format("02.01.2006")))

	for _, h := range horoscopes {
		sb.WriteString(fmt.Sprintf("<b>%s</b>: %s\n\n", h.ZodiacSign, h.Content))
	}

	sb.WriteString("#гороскоп #текстовыйгороскоп")

	return c.telegramClient.ReplyMessage(update.Message.Chat.ID, update.Message.ID, sb.String())
}
