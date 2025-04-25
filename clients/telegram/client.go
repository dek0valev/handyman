package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	hc       *http.Client
	baseURL  string
	basePath string
}

func NewClient(authToken string) *Client {
	return &Client{
		hc:       &http.Client{},
		baseURL:  "https://api.telegram.org",
		basePath: "/bot" + authToken,
	}
}

func (c *Client) GetUpdates(offset, limit int64) ([]Update, error) {
	const op = "clients.telegram.Updates"

	q := url.Values{}
	q.Set("offset", strconv.FormatInt(offset, 10))
	q.Set("limit", strconv.FormatInt(limit, 10))

	req, err := http.NewRequest(http.MethodGet, c.baseURL+c.basePath+"/getUpdates?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result.Result, nil
}

func (c *Client) SendMessage(chatID int64, text string) error {
	const op = "clients.telegram.SendMessage"

	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.basePath+"/sendMessage", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	return nil
}

func (c *Client) ReplyMessage(chatID int64, replyMessageID int64, text string) error {
	const op = "clients.telegram.ReplyMessage"

	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
		"reply_parameters": map[string]any{
			"message_id": replyMessageID,
		},
		"parse_mode": "HTML",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.basePath+"/sendMessage", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	return nil
}

func (c *Client) GetFile(fileID string) (FileInfo, error) {
	const op = "clients.telegram.GetFile"

	q := url.Values{}
	q.Set("file_id", fileID)

	req, err := http.NewRequest(http.MethodGet, c.baseURL+c.basePath+"/getFile?"+q.Encode(), nil)
	if err != nil {
		return FileInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return FileInfo{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return FileInfo{}, fmt.Errorf("%s: %w", op, err)
		}

		return FileInfo{}, fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return FileInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result FileInfo `json:"result"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return FileInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return result.Result, nil
}

func (c *Client) DownloadFile(filePath string) ([]byte, error) {
	const op = "clients.telegram.DownloadFile"

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	urlPath, err := url.JoinPath(u.Path, "file", c.basePath, filePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	u.Path = urlPath

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: unexpected status code %d: %s", op, resp.StatusCode, bodyBytes)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return bodyBytes, nil
}
