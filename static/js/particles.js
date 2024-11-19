document.addEventListener('DOMContentLoaded', function() {
    const particlesContainer = document.getElementById('particles');
    const particleCount = 100;

    for (let i = 0; i < particleCount; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle';
        
        // Более разнообразные размеры частиц
        const size = Math.random() * 8 + 2 + 'px';
        particle.style.width = size;
        particle.style.height = size;
        
        // Случайные позиции
        particle.style.left = Math.random() * 100 + 'vw';
        particle.style.top = Math.random() * 100 + 'vh';
        
        // Случайная задержка и прозрачность
        particle.style.animationDelay = Math.random() * 15 + 's';
        particle.style.opacity = Math.random() * 0.5 + 0.3;

        particlesContainer.appendChild(particle);
    }
}); 