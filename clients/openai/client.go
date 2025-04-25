package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Client struct {
	hc       *http.Client
	apiKey   string
	baseURL  string
	basePath string
}

func NewClient(apiKey string) *Client {
	return &Client{
		hc:       &http.Client{},
		apiKey:   apiKey,
		baseURL:  "https://api.openai.com",
		basePath: "/v1",
	}
}

func (c *Client) GenerateResponse(instructions, content string) (string, error) {
	const op = "clients.openai.GenerateResponse"

	payload := map[string]any{
		"model":        "gpt-4o-mini",
		"instructions": instructions,
		"input":        content,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.basePath+"/responses", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		return "", fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var result struct {
		ID     string `json:"id"`
		Output []struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			Status  string `json:"status"`
			Content []struct {
				Type        string `json:"type"`
				Annotations []any  `json:"annotations"`
				Text        string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"output"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result.Output[0].Content[0].Text, nil
}

func (c *Client) TranscribeAudio(audioFileBytes []byte, fileName string) (string, error) {
	const op = "clients.openai.TranscribeAudio"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if _, err := io.Copy(part, bytes.NewReader(audioFileBytes)); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := writer.WriteField("model", "gpt-4o-mini-transcribe"); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.basePath+"/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		return "", fmt.Errorf("%s: %s: %s", op, resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var result struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result.Text, nil
}
