async function redirectToTopicWiki(category) {
    try {
        const button = event.target;
        const originalText = button.textContent;
        button.disabled = true;
        button.classList.add('loading');
        
        const response = await fetch(`/api/random-article?category=${category}`);
        const data = await response.json();
        
        if (data.url) {
            // Показываем модальное окно с кратким содержанием
            if (data.summary) {
                showSummaryModal(data.title, data.summary, data.url);
            } else {
                window.location.href = data.url;
            }
        } else {
            throw new Error('No URL in response');
        }
        
        button.disabled = false;
        button.classList.remove('loading');
        button.textContent = originalText;
    } catch (error) {
        console.error('Ошибка:', error);
        // ... обработка ошибки ...
    }
}

function showSummaryModal(title, summary, url) {
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    modal.innerHTML = `
        <div class="summary-content">
            <h2>${title}</h2>
            <p>${summary}</p>
            <div class="summary-buttons">
                <button onclick="window.location.href='${url}'">Читать статью</button>
                <button onclick="this.closest('.summary-modal').remove()">Закрыть</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}

function getButtonText(category) {
    const texts = {
        'science': 'Наука',
        'it': 'IT',
        'sport': 'Спорт',
        'books': 'Книги',
        'games': 'Игры',
        'movies': 'Фильмы/Сериалы'
    };
    return texts[category] || category;
}

async function searchWiki() {
    const searchInput = document.getElementById('searchInput');
    const query = searchInput.value.trim();
    
    if (!query) {
        return;
    }

    const button = document.querySelector('.search-button');
    button.disabled = true;
    button.textContent = 'Поиск...';

    try {
        const response = await fetch(`/api/search-wiki?query=${encodeURIComponent(query)}`);
        
        // Проверяем статус ответа
        if (!response.ok) {
            throw new Error('Ошибка сервера');
        }

        const data = await response.json();
        
        if (data.notFound) {
            showNotFoundModal();
        } else if (data.url) {
            showSummaryModalWithSimilar(data.title, data.summary, data.url, data.similar);
        } else {
            showErrorModal('Не удалось получить результаты поиска');
        }
        
    } catch (error) {
        console.error('Ошибка:', error);
        showErrorModal('Произошла ошибка при поиске. Пожалуйста, попробуйте позже.');
    } finally {
        button.disabled = false;
        button.textContent = 'Найти';
    }
}

function showSummaryModalWithSimilar(title, summary, url, similar, isFromTopic = false) {
    console.log('Показываем модальное окно:', { title, summary, url, similar, isFromTopic });
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    
    // Формируем HTML для похожих статей
    const similarArticlesHTML = similar && Array.isArray(similar) && similar.length > 0 
        ? `
            <div class="similar-articles">
                <h3>Похожие статьи:</h3>
                <ul>
                    ${similar.map(article => {
                        const articleTitle = typeof article === 'string' ? article : article.title;
                        const articleUrl = typeof article === 'string' 
                            ? `https://ru.wikipedia.org/wiki/${encodeURIComponent(articleTitle)}`
                            : article.url;
                        return `<li><a href="${articleUrl}" target="_blank">${articleTitle}</a></li>`;
                    }).join('')}
                </ul>
            </div>`
        : '';

    modal.innerHTML = `
        <div class="summary-content">
            <h2>${title}</h2>
            <p>${summary}</p>
            ${similarArticlesHTML}
            <div class="summary-buttons">
                <button onclick="window.open('${url}', '_blank')">Читать полностью</button>
                <button onclick="this.closest('.summary-modal').remove()">Закрыть</button>
                ${isFromTopic ? `<button onclick="getNewArticleForTopic('${currentTopic}')">Новая статья на эту тему</button>` : ''}
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}

function showNotFoundModal() {
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    modal.innerHTML = `
        <div class="summary-content not-found-content">
            <div class="not-found-icon">❌</div>
            <h2>Статья не найдена</h2>
            <p>К сожалению, по вашему запросу ничего не найдено. Попробуйте изменить поисковый запрос.</p>
            <div class="summary-buttons">
                <button class="retry-button" onclick="this.closest('.summary-modal').remove()">Попробовать снова</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}

// Добавляем обработчик Enter для поискового поля
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('searchInput');
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            searchWiki();
        }
    });
});

// Объединить функции показа модальных окон
function showModal(content) {
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    modal.innerHTML = content;
    document.body.appendChild(modal);
}

function getSimilarArticlesHTML(similarTitles) {
    if (similarTitles.length > 0) {
        return `
            <div class="similar-articles">
                <h3>Похожие статьи:</h3>
                <ul>
                    ${similarTitles.map(title => `
                        <li>
                            <a href="https://ru.wikipedia.org/wiki/${encodeURIComponent(title)}" 
                               target="_blank">${title}</a>
                        </li>
                    `).join('')}
                </ul>
            </div>
        `;
    } else {
        return '';
    }
}

// Добавляем ноую функцию для отображения ошибок
function showErrorModal(message) {
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    modal.innerHTML = `
        <div class="summary-content error-content">
            <div class="error-icon">⚠️</div>
            <h2>Ошибка</h2>
            <p>${message}</p>
            <div class="summary-buttons">
                <button onclick="this.closest('.summary-modal').remove()">Закрыть</button>
                <button onclick="this.closest('.summary-modal').remove(); redirectToTopicWiki('${currentTopic}')">Попробовать снова</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}

// Добавим глобальную переменную для хранения текущего топика
let currentTopic = '';

// Обновим функцию redirectToTopicWiki
function redirectToTopicWiki(topic) {
    console.log('Выбран топик:', topic);
    currentTopic = topic;
    
    fetch(`/api/random-article?category=${encodeURIComponent(topic)}`)
        .then(response => {
            console.log('Получен ответ:', response.status);
            return response.json();
        })
        .then(data => {
            console.log('Получены данные:', data);
            if (data.error) {
                throw new Error(data.error);
            }
            if (!data.url || !data.title) {
                throw new Error('Неполные данные от сервера');
            }
            showSummaryModalWithSimilar(
                data.title,
                data.summary || 'Краткое содержание недоступно',
                data.url,
                data.similar || [],
                true // Указываем, что это вызов из топика
            );
        })
        .catch(error => {
            console.error('Ошибка:', error);
            showErrorModal(`Не удалось получить статью: ${error.message}`);
        });
}

// Функция для получения новой статьи по текущей теме
function getNewArticleForTopic(topic) {
    console.log('Запрос новой статьи для топика:', topic);
    const currentModal = document.querySelector('.summary-modal');
    if (currentModal) {
        currentModal.remove();
    }
    redirectToTopicWiki(topic);
}
