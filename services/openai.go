package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	MaxTokens int      `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// CallChatGPT - تماس با API ChatGPT
func CallChatGPT(apiKey, systemPrompt, userMessage string, isVIP bool) (string, int, error) {
	// انتخاب مدل بر اساس وضعیت کاربر
	model := selectModel(isVIP)

	// آماده‌سازی درخواست
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	requestBody := ChatGPTRequest{
		Model:    model,
		Messages: messages,
		MaxTokens: 2000, // محدودیت توکن برای جلوگیری از هزینه‌های بالا
	}

	// تبدیل به JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, fmt.Errorf("خطا در آماده‌سازی درخواست: %v", err)
	}

	// ایجاد درخواست HTTP
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("خطا در ایجاد درخواست: %v", err)
	}

	// تنظیم هدرها
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// ارسال درخواست
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("خطا در ارسال درخواست به OpenAI: %v", err)
	}
	defer resp.Body.Close()

	// خواندن پاسخ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("خطا در خواندن پاسخ: %v", err)
	}

	// پردازش پاسخ
	var chatResponse ChatGPTResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return "", 0, fmt.Errorf("خطا در پردازش پاسخ JSON: %v", err)
	}

	// بررسی خطاهای API
	if chatResponse.Error.Message != "" {
		return "", 0, fmt.Errorf("خطای OpenAI: %s", chatResponse.Error.Message)
	}

	if len(chatResponse.Choices) == 0 {
		return "", 0, fmt.Errorf("پاسخی از OpenAI دریافت نشد")
	}

	// بررسی finish_reason
	if chatResponse.Choices[0].FinishReason == "length" {
		return chatResponse.Choices[0].Message.Content + "\n\n⚠️ پاسخ به دلیل محدودیت توکن قطع شد.", 
		       chatResponse.Usage.TotalTokens, nil
	}

	return chatResponse.Choices[0].Message.Content, chatResponse.Usage.TotalTokens, nil
}

// انتخاب مدل بر اساس وضعیت کاربر
func selectModel(isVIP bool) string {
	if isVIP {
		// کاربران VIP می‌توانند از GPT-4 استفاده کنند
		return "gpt-4"
	}
	// کاربران عادی از GPT-3.5 استفاده می‌کنند
	return "gpt-3.5-turbo"
}

// بررسی اعتبار API Key
func ValidateAPIKey(apiKey string) (bool, error) {
	// درخواست ساده برای بررسی اعتبار API Key
	messages := []Message{
		{
			Role:    "user",
			Content: "سلام",
		},
	}

	requestBody := ChatGPTRequest{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
		MaxTokens: 5,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// اگر status code 200 باشد، API Key معتبر است
	if resp.StatusCode == 200 {
		return true, nil
	}

	// خواندن پیام خطا برای تشخیص بهتر
	body, _ := io.ReadAll(resp.Body)
	var errorResponse struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	
	json.Unmarshal(body, &errorResponse)
	
	if strings.Contains(errorResponse.Error.Message, "invalid_api_key") {
		return false, fmt.Errorf("API Key نامعتبر است")
	} else if strings.Contains(errorResponse.Error.Message, "insufficient_quota") {
		return false, fmt.Errorf("سقف مصرف API Key به پایان رسیده است")
	}

	return false, fmt.Errorf("خطا در بررسی API Key: %s", errorResponse.Error.Message)
}

// محاسبه هزینه بر اساس تعداد توکن
func CalculateCost(tokens int, isVIP bool) float64 {
	var costPerToken float64
	
	if isVIP {
		// هزینه برای GPT-4 (تقریبی)
		costPerToken = 0.03 / 1000 // 0.03 دلار per 1K tokens
	} else {
		// هزینه برای GPT-3.5-turbo (تقریبی)
		costPerToken = 0.002 / 1000 // 0.002 دلار per 1K tokens
	}

	return float64(tokens) * costPerToken
}

// تولید محتوا برای کانال‌ها (ویژه کاربران VIP)
func GenerateChannelContent(apiKey, prompt string, wordCount int) (string, int, error) {
	systemPrompt := fmt.Sprintf(
		"تو یک تولیدکننده محتوای حرفه‌ای هستی. محتوایی تولید کن که:\n"+
		"- جذاب و مفید باشد\n"+
		- حدود %d کلمه باشد\n"+
		"- برای انتشار در کانال‌های تلگرام مناسب باشد\n"+
		"- دارای ساختار منظم و پاراگراف‌بندی مناسب\n"+
		"- حاوی نکات کاربردی و ارزشمند",
		wordCount,
	)

	userMessage := fmt.Sprintf(
		"لطفاً در مورد موضوع زیر محتوا تولید کن:\n%s",
		prompt,
	)

	return CallChatGPT(apiKey, systemPrompt, userMessage, true) // همیشه از مدل VIP استفاده می‌شود
}

// خلاصه‌سازی پاسخ برای گروه‌ها در صورت طولانی بودن
func SummarizeResponse(response string, maxLength int) string {
	if len(response) <= maxLength {
		return response
	}

	// کوتاه کردن پاسخ و اضافه کردن نشانه
	shortened := response[:maxLength-10] + "..."
	return shortened
}

// لاگ کردن درخواست‌ها برای دیباگ
func LogAPIRequest(apiKeyLast4, model, prompt string, tokens int, success bool) {
	status := "✅ موفق"
	if !success {
		status = "❌ ناموفق"
	}

	log.Printf("درخواست API - Key: %s... | Model: %s | Tokens: %d | Status: %s", 
		apiKeyLast4, model, tokens, status)
}

// مدیریت خطاهای رایج OpenAI
func HandleOpenAIError(err error) string {
	errMsg := err.Error()
	
	switch {
	case strings.Contains(errMsg, "invalid_api_key"):
		return "❌ API Key نامعتبر است. لطفاً API Key خود را بررسی کنید."
	
	case strings.Contains(errMsg, "insufficient_quota"):
		return "❌ سقف مصرف API Key شما به پایان رسیده است. لطفاً API Key جدیدی اضافه کنید."
	
	case strings.Contains(errMsg, "rate_limit"):
		return "⏳ محدودیت نرخ درخواست. لطفاً چند لحظه صبر کنید."
	
	case strings.Contains(errMsg, "context_length"):
		return "📝 پیام بسیار طولانی است. لطفاً سوال خود را کوتاه‌تر کنید."
	
	case strings.Contains(errMsg, "timeout"):
		return "⏰ زمان درخواست به پایان رسید. لطفاً مجدد تلاش کنید."
	
	default:
		return "❌ خطا در ارتباط با سرویس ChatGPT. لطفاً مجدد تلاش کنید."
	}
}
