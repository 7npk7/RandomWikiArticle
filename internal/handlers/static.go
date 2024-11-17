package handlers

import (
	"io"
	"net/http"
)

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>` + getIndexHTML()
	io.WriteString(w, html)
}

func getIndexHTML() string {
	return `
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <meta http-equiv="Pragma" content="no-cache">
    <meta http-equiv="Expires" content="0">
    <title>Тематическая Википедия</title>
    <link rel="stylesheet" href="/static/css/style.css?v=4">
</head>
<body>
    <div class="container">
        <h1>Выберите топик для случайной статьи на Википедии</h1>
        <div class="topics-grid">
            <button onclick="redirectToTopicWiki('science')">Наука</button>
            <button onclick="redirectToTopicWiki('it')">IT</button>
            <button onclick="redirectToTopicWiki('sport')">Спорт</button>
            <button onclick="redirectToTopicWiki('books')">Книги</button>
            <button onclick="redirectToTopicWiki('games')">Игры</button>
            <button onclick="redirectToTopicWiki('movies')">Фильмы/Сериалы</button>
        </div>
    </div>
    <script src="/static/js/script.js"></script>
</body>
</html>
`
}
