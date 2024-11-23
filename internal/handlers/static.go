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

func getParticlesHTML() string {
	particles := ""
	for i := 0; i < 2000; i++ {
		particles += `<div class="particle"></div>`
	}
	return particles
}

func getIndexHTML() string {
	return `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <meta http-equiv="Pragma" content="no-cache">
    <meta http-equiv="Expires" content="0">
    <title>Тематическая Википедия</title>
    
    <!-- Favicon -->
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🌐</text></svg>">
    
    <!-- Альтернативные иконки для разных платформ -->
    <link rel="apple-touch-icon" sizes="180x180" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🌐</text></svg>">
    <link rel="icon" type="image/svg+xml" sizes="32x32" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🌐</text></svg>">
    <link rel="icon" type="image/svg+xml" sizes="16x16" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🌐</text></svg>">
    
    <link rel="stylesheet" href="/static/css/style.css?v=4">

    <!-- Yandex.Metrika counter -->
    <script type="text/javascript" >
       (function(m,e,t,r,i,k,a){m[i]=m[i]||function(){(m[i].a=m[i].a||[]).push(arguments)};
       m[i].l=1*new Date();
       for (var j = 0; j < document.scripts.length; j++) {if (document.scripts[j].src === r) { return; }}
       k=e.createElement(t),a=e.getElementsByTagName(t)[0],k.async=1,k.src=r,a.parentNode.insertBefore(k,a)})
       (window, document, "script", "https://mc.yandex.ru/metrika/tag.js", "ym");

       ym(98991910, "init", {
            clickmap:true,
            trackLinks:true,
            accurateTrackBounce:true
       });
    </script>
    <noscript><div><img src="https://mc.yandex.ru/watch/98991910" style="position:absolute; left:-9999px;" alt="" /></div></noscript>
    <!-- /Yandex.Metrika counter -->
</head>
<body>
    <div class="particles">
        ` + getParticlesHTML() + `
    </div>
    <div class="glow"></div>
    <div class="glow"></div>
    <div class="glow"></div>
    
    <div class="container">
        <div class="search-container">
            <input type="text" id="searchInput" placeholder="Поиск статьи..." class="search-input">
            <button onclick="searchWiki()" class="search-button">Найти</button>
        </div>

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
