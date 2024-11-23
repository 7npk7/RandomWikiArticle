package handlers

import (
	"CursorWebApp/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

type WikiResponse struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"search"`
		CategoryMembers []struct {
			Title string `json:"title"`
		} `json:"categorymembers"`
	} `json:"query"`
	Continue struct {
		CmContinue string `json:"cmcontinue"`
	} `json:"continue"`
}

type ArticleResponse struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
}

// Обновляем карту соответствий категорий
var categoryMap = map[string]string{
	"science": "Наука",
	"it":      "Информационные_технологии",
	"sport":   "Спорт",
	"books":   "Литература",
	"games":   "Компьютерные_игры",
	"movies":  "Кинематограф",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandomArticleHandler(w http.ResponseWriter, r *http.Request) {
	// Включаем CORS и устанавливаем заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Получаем категорию из параметров запроса
	category := r.URL.Query().Get("category")
	fmt.Printf("🔍 Получен запрос для категории: %s\n", category)

	if category == "" {
		http.Error(w, `{"error": "Category is required"}`, http.StatusBadRequest)
		return
	}

	// Получаем русское название категории
	ruCategory, ok := categoryMap[category]
	if !ok {
		fmt.Printf("❌ Неизвестная категория: %s\n", category)
		http.Error(w, `{"error": "Invalid category"}`, http.StatusBadRequest)
		return
	}

	// Формируем корректный URL для API запроса
	apiURL := fmt.Sprintf(
		"https://ru.wikipedia.org/w/api.php?action=query&list=categorymembers&cmtitle=%s&format=json&cmlimit=500&cmnamespace=0&origin=*",
		url.QueryEscape("Категория:"+ruCategory),
	)
	fmt.Printf("📌 URL запроса: %s\n", apiURL)

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{Timeout: 10 * time.Second}

	// Выполняем запрос к API Wikipedia
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create request"})
		return
	}

	// Устанавливаем правильные заголовки
	req.Header.Set("User-Agent", "WikiRandomArticle/1.0 (https://your-domain.com; your@email.com)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Ошибка выполнения запроса: %v\n", err)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch articles"})
		return
	}
	defer resp.Body.Close()

	// Читаем тело ответа для отладки
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения ответа: %v\n", err)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read response"})
		return
	}

	// Выводим первые 200 символов ответа для отладки
	fmt.Printf("📝 Начало ответа: %s\n", string(body[:min(len(body), 200)]))

	var result struct {
		Query struct {
			CategoryMembers []struct {
				Title string `json:"title"`
			} `json:"categorymembers"`
		} `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ Ошибка декодирования JSON: %v\n", err)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode Wikipedia response"})
		return
	}

	if len(result.Query.CategoryMembers) == 0 {
		fmt.Printf("❌ Статьи не найдены для категории: %s\n", ruCategory)
		json.NewEncoder(w).Encode(map[string]string{"error": "No articles found"})
		return
	}

	// Выбираем случайную статью
	randomArticle := result.Query.CategoryMembers[rand.Intn(len(result.Query.CategoryMembers))]

	// Формируем URL статьи
	articleURL := fmt.Sprintf("https://ru.wikipedia.org/wiki/%s",
		url.PathEscape(randomArticle.Title))

	// Формируем ответ
	response := ArticleResponse{
		URL:   articleURL,
		Title: randomArticle.Title,
	}

	// Пытаемся получить содержание статьи
	if content, err := getArticleContent(randomArticle.Title); err == nil && content != "" {
		if gigaChatService := services.NewGigaChatService(os.Getenv("GIGACHAT_TOKEN")); gigaChatService != nil {
			if summary, err := gigaChatService.GenerateSummary(r.Context(), content); err == nil {
				response.Summary = summary
			}
		}
	}

	// Отправляем ответ
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("❌ Ошибка отправки ответа: %v\n", err)
		http.Error(w, `{"error": "Failed to send response"}`, http.StatusInternalServerError)
		return
	}
}

func getArticleContent(title string) (string, error) {
	apiURL := fmt.Sprintf(
		"https://ru.wikipedia.org/w/api.php?action=query&prop=extracts&exintro=true&format=json&titles=%s",
		url.QueryEscape(title),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Извлекаем текст из ответа API
	if pages, ok := result["query"].(map[string]interface{})["pages"].(map[string]interface{}); ok {
		for _, page := range pages {
			if pageMap, ok := page.(map[string]interface{}); ok {
				if extract, ok := pageMap["extract"].(string); ok {
					return extract, nil
				}
			}
		}
	}

	return "", fmt.Errorf("не удалось плучить содержание статьи")
}

// Вспомогательная функция для отправки ошибок
func sendError(w http.ResponseWriter, message string, status int) {
	fmt.Printf("❌ %s (статус: %d)\n", message, status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// Добавим новую структуру для ответа
type SearchResponse struct {
	URL      string   `json:"url,omitempty"`
	Title    string   `json:"title,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Similar  []string `json:"similar,omitempty"` // Добавляем поле для похожих статей
	NotFound bool     `json:"notFound,omitempty"`
}

func SearchWikiHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		sendError(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Добавляем отладочный вывод
	fmt.Printf("Поисковый запрос: %s\n", query)

	// Формируем URL для поиска с дополнительными параметрами
	apiURL := fmt.Sprintf(
		"https://ru.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&srlimit=5&utf8=1&srwhat=text&srenablerewrites=1",
		url.QueryEscape(query),
	)

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// Добавляем User-Agent
	req.Header.Set("User-Agent", "WikiSearchApp/1.0")

	resp, err := client.Do(req)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to search: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Читаем тело ответа для отладки
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	// Выводим ответ API для отладки
	fmt.Printf("Ответ API Wikipedia: %s\n", string(body))

	var result struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
			} `json:"search"`
			SearchInfo struct {
				TotalHits int `json:"totalhits"`
			} `json:"searchinfo"`
		} `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		sendError(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	// Проверяем наличие результатов
	if len(result.Query.Search) == 0 {
		// Пробуем поиск с исправлением опечаток
		apiURL = fmt.Sprintf(
			"https://ru.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&srlimit=5&utf8=1&srwhat=text&srenablerewrites=1&srinfo=suggestion",
			url.QueryEscape(query),
		)

		req, _ = http.NewRequest("GET", apiURL, nil)
		req.Header.Set("User-Agent", "WikiSearchApp/1.0")

		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SearchResponse{
				NotFound: true,
			})
			return
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SearchResponse{
				NotFound: true,
			})
			return
		}
	}

	if len(result.Query.Search) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SearchResponse{
			NotFound: true,
		})
		return
	}

	// Получаем основную статью
	mainArticle := result.Query.Search[0]

	// Собираем похожие статьи
	var similarTitles []string
	for i := 1; i < len(result.Query.Search); i++ {
		similarTitles = append(similarTitles, result.Query.Search[i].Title)
	}

	// Получаем содержание основной статьи
	content, err := getArticleContent(mainArticle.Title)
	var summary string
	if err == nil && content != "" {
		gigaChatService := services.NewGigaChatService(os.Getenv("GIGACHAT_TOKEN"))
		if gigaChatService != nil {
			summary, _ = gigaChatService.GenerateSummary(r.Context(), content)
		}
	}

	articleURL := fmt.Sprintf("https://ru.wikipedia.org/wiki/%s",
		url.PathEscape(mainArticle.Title))

	// Добавляем отладочный вывод
	fmt.Printf("Найдена статья: %s\nПохожие статьи: %v\n", mainArticle.Title, similarTitles)

	response := SearchResponse{
		URL:     articleURL,
		Title:   mainArticle.Title,
		Summary: summary,
		Similar: similarTitles,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Ошибка при кодировании ответа: %v\n", err)
	}
}

// Объединить общую логику поиска
