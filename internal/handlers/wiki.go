package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type WikiResponse struct {
	Continue struct {
		CmContinue string `json:"cmcontinue"`
	} `json:"continue"`
	Query struct {
		CategoryMembers []struct {
			Title string `json:"title"`
		} `json:"categorymembers"`
	} `json:"query"`
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

	// Возвращаем успешный результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":   articleURL,
		"title": randomArticle.Title,
	})
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
