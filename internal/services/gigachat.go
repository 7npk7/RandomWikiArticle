package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type GigaChatService struct {
	token  string
	client *http.Client
}

func NewGigaChatService(token string) *GigaChatService {
	if token == "" {
		return nil
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &GigaChatService{
		token:  token,
		client: client,
	}
}

func (s *GigaChatService) getAuthToken() (string, error) {
	authURL := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

	payload := "scope=GIGACHAT_API_PERS"

	req, err := http.NewRequest("POST", authURL, bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	req.Header.Set("RqUID", uuid.New().String())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Access_token string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("неверный код ответа: %d", resp.StatusCode)
	}

	return result.Access_token, nil
}

func (s *GigaChatService) GenerateSummary(ctx context.Context, text string) (string, error) {
	accessToken, err := s.getAuthToken()
	if err != nil {
		return "", fmt.Errorf("не удалось получить токен доступа: %w", err)
	}

	url := "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "GigaChat:latest",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("Создай краткое содержание следующего текста на русском языке. Используй не более 3-4 предложений: %s", text),
			},
		},
		"temperature": 0.3,
		"max_tokens":  500,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	// Отладочный вывод в более читаемом формате
	fmt.Printf("Заголовки запроса:\n")
	for key, values := range req.Header {
		fmt.Printf("  %s: %s\n", key, values)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	// Отладочный вывод в более читаемом формате
	fmt.Printf("Ответ API:\n")
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))

	// Извлекаем ответ из результата
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("неверный формат ответа API")
}

func (s *GigaChatService) GetSimilarArticles(ctx context.Context, title string, content string) ([]string, error) {
	accessToken, err := s.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить токен доступа: %w", err)
	}

	url := "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"

	prompt := fmt.Sprintf(`На основе статьи "%s" с содержанием: "%s" 
		предложи 3-4 темы для похожих статей. Ответ дай в формате простого списка тем, 
		каждая тема с новой строки, без нумерации и дополнительных пояснений.`, title, content)

	payload := map[string]interface{}{
		"model": "GigaChat:latest",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  200,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	// Извлекаем ответ и разбиваем его на отдельные темы
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					// Разбиваем ответ на строки и очищаем их
					topics := strings.Split(content, "\n")
					var cleanTopics []string
					for _, topic := range topics {
						topic = strings.TrimSpace(topic)
						if topic != "" {
							cleanTopics = append(cleanTopics, topic)
						}
					}
					return cleanTopics, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("неверный формат ответа API")
}
