package main

import (
	"CursorWebApp/internal/handlers"
	"fmt"
	"net/http"
)

func main() {
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

	fmt.Println("Сервер запущен на порту 8080...")
	http.ListenAndServe(":8080", nil)
}
