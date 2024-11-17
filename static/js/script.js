async function redirectToTopicWiki(category) {
    try {
        const button = event.target;
        const originalText = button.textContent;
        button.disabled = true;
        
        const timestamp = new Date().getTime();
        const response = await fetch(`/api/random-article?category=${category}&t=${timestamp}`, {
            headers: {
                'Cache-Control': 'no-cache',
                'Pragma': 'no-cache'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.url) {
            button.disabled = false;
            button.textContent = originalText;
            window.location.href = data.url;
        } else {
            throw new Error('No URL in response');
        }
    } catch (error) {
        console.error('Произошла ошибка при получении случайной статьи:', error);
        const button = event.target;
        button.textContent = 'Ошибка! Попробуйте снова';
        button.classList.add('error');
        setTimeout(() => {
            button.disabled = false;
            button.textContent = getButtonText(category);
            button.classList.remove('error');
        }, 2000);
    }
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
