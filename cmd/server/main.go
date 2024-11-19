package main

import (
	"CursorWebApp/internal/handlers"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Пробуем загрузить .env из разных возможных расположений
	envPaths := []string{
		".env",
		"../../.env",
		"../../../.env",
	}

	var loaded bool
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loaded = true
			fmt.Printf("✅ Загружен .env файл из: %s\n", path)
			break
		}
	}

	if !loaded {
		fmt.Println("⚠️ Не удалось загрузить .env файл")
	}
}

func main() {

	if os.Getenv("GIGACHAT_TOKEN") == "" {
		fmt.Println("⚠️ GIGACHAT_TOKEN не установлен в переменных окружения")
	}

	// Настройка обработчика статических файлов с отключением кэширования
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "-1")
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	}))

	// Обработчики
	http.HandleFunc("/", handlers.StaticHandler)
	http.HandleFunc("/api/random-article", handlers.GetRandomArticleHandler)
	http.HandleFunc("/api/search-wiki", handlers.SearchWikiHandler)

	fmt.Println("Сервер запущен на порту 8080...")
	http.ListenAndServe(":8080", nil)
}
