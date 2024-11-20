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

// Добавляем карту соответствий в начало файла после объявления структур
var categoryMap = map[string]string{
	"science": "Наука",
	"it":      "Информатика",
	"sport":   "Спорт",
	"books":   "Литература",
	"games":   "Компьютерные_игры",
	"movies":  "Кинематограф",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandomArticleHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	fmt.Printf("🔍 Запрос статьи из категории: %s\n", category)

	if category == "" {
		sendError(w, "Category is required", http.StatusBadRequest)
		return
	}

	// Получаем русское название категории
	wikiCategory, exists := categoryMap[category]
	if !exists {
		sendError(w, "Invalid category", http.StatusBadRequest)
		return
	}

	// Формируем URL для API запроса с добавлением случайного параметра
	categoryPath := "Категория:" + wikiCategory
	apiURL := fmt.Sprintf(
		"https://ru.wikipedia.org/w/api.php?action=query&list=categorymembers&cmtitle=%s&format=json&cmlimit=500&cmnamespace=0&cmtype=page",
		url.QueryEscape(categoryPath),
	)

	fmt.Printf("Запрашиваемый URL: %s\n", apiURL)

	// Добавляем заголовки для предотвращения кэширования
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("User-Agent", "RandomWikiArticle/1.0 (educational project)")

	var allArticles []struct {
		Title string `json:"title"`
	}
	cmcontinue := ""

	// Получаем до 2000 статей (4 запроса по 500)
	for i := 0; i < 4; i++ {
		apiURL := fmt.Sprintf(
			"https://ru.wikipedia.org/w/api.php?action=query&list=categorymembers&cmtitle=%s&format=json&cmlimit=500&cmnamespace=0",
			url.QueryEscape(categoryPath),
		)

		if cmcontinue != "" {
			apiURL += "&cmcontinue=" + url.QueryEscape(cmcontinue)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Ошибка API запроса: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		var wikiResp WikiResponse
		if err := json.NewDecoder(resp.Body).Decode(&wikiResp); err != nil {
			fmt.Printf("Ошибка декодирования ответа: %v\n", err)
			continue
		}

		allArticles = append(allArticles, wikiResp.Query.CategoryMembers...)

		if wikiResp.Continue.CmContinue == "" {
			break
		}
		cmcontinue = wikiResp.Continue.CmContinue
	}

	fmt.Printf("✅ Найдено статей: %d\n", len(allArticles))

	if len(allArticles) == 0 {
		sendError(w, fmt.Sprintf("No articles found in category: %s", category), http.StatusNotFound)
		return
	}

	// Выбираем случайную статью из всего списка
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)
	randomArticle := allArticles[rnd.Intn(len(allArticles))]

	if randomArticle.Title == "" {
		sendError(w, "Failed to get valid random article", http.StatusInternalServerError)
		return
	}

	fmt.Printf("📎 Выбрана статья: %s\n", randomArticle.Title)

	articleURL := fmt.Sprintf("https://ru.wikipedia.org/wiki/%s",
		url.PathEscape(randomArticle.Title))

	// Получаем содержание статьи
	articleContent, err := getArticleContent(randomArticle.Title)
	if err != nil {
		fmt.Printf("Ошибка получения содержания статьи: %v\n", err)
	}

	fmt.Printf("📄 Получено содержание статьи длиной: %d символов\n", len(articleContent))

	// Генерируем краткое содержание через GigaChat
	var summary string
	gigaChatService := services.NewGigaChatService(os.Getenv("GIGACHAT_TOKEN"))
	if gigaChatService == nil {
		fmt.Printf("❌ Не удалось инициализировать GigaChat сервис\n")
	} else {
		summary, err = gigaChatService.GenerateSummary(r.Context(), articleContent)
		if err != nil {
			fmt.Printf("❌ Ошибка генерации краткого содержания: %v\n", err)
		}
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ArticleResponse{
		URL:     articleURL,
		Title:   randomArticle.Title,
		Summary: summary,
	})
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
