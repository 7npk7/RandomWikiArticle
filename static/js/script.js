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

    try {
        const button = document.querySelector('.search-button');
        const originalText = button.textContent;
        button.disabled = true;
        button.textContent = 'Поиск...';
        
        const response = await fetch(`/api/search-wiki?query=${encodeURIComponent(query)}`);
        const data = await response.json();
        
        if (data.notFound) {
            showNotFoundModal();
        } else if (data.url) {
            showSummaryModalWithSimilar(data.title, data.summary, data.url, data.similar);
        }
        
        button.disabled = false;
        button.textContent = originalText;
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Произошла ошибка при поиске');
    }
}

function showSummaryModalWithSimilar(title, summary, url, similarTitles) {
    const content = `
        <div class="summary-content">
            <h2>${title}</h2>
            <p>${summary || 'Краткое содержание недоступно'}</p>
            ${getSimilarArticlesHTML(similarTitles)}
            <div class="summary-buttons">
                <button onclick="window.location.href='${url}'">Читать статью</button>
                <button onclick="this.closest('.summary-modal').remove()">Закрыть</button>
            </div>
        </div>
    `;
    showModal(content);
}

function showNotFoundModal() {
    const modal = document.createElement('div');
    modal.className = 'summary-modal';
    modal.innerHTML = `
        <div class="summary-content">
            <h2>Статья не найдена</h2>
            <p>К сожалению, по вашему запросу ничего не найдено.</p>
            <div class="summary-buttons">
                <button onclick="this.closest('.summary-modal').remove()">Закрыть</button>
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
