package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChatRequest برای ارسال درخواست به OpenAI
type ChatRequest struct {
	Model    string              `json:"model"`
	Messages []map[string]string `json:"messages"`
}

// ChatResponse ساختار پاسخ API
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// SendChatWithKey — ارسال درخواست ChatGPT با کلید اختصاصی کاربر
func SendChatWithKey(apiKey, model, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := ChatRequest{
		Model: model,
		Messages: []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 40 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("خطا در ارسال درخواست: %v", err)
	}
	defer res.Body.Close()

	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("OpenAI پاسخ خطا داد (%d): %s", res.StatusCode, string(respBody))
	}

	var parsed ChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("خطا در پردازش پاسخ: %v", err)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("پاسخی از GPT دریافت نشد")
	}

	return parsed.Choices[0].Message.Content, nil
}
