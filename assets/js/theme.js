document.addEventListener('DOMContentLoaded', () => {
    const toggle = document.getElementById('theme-toggle');
    const html = document.documentElement;
    
    // If toggle doesn't exist (user not logged in), skip
    if (!toggle) {
        return;
    }

    let currentTheme = html.getAttribute('data-bs-theme');
    const icon = toggle.querySelector('i');
    
    // Set initial icon based on current theme
    updateIcon(currentTheme);

    toggle.addEventListener('click', async (e) => {
        e.preventDefault();
        
        // Determine new theme
        let newTheme;
        if (currentTheme === 'dark') {
            newTheme = 'light';
        } else if (currentTheme === 'light') {
            newTheme = 'dark';
        } else {
            // If auto, check system preference or default to dark
            const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            newTheme = systemPrefersDark ? 'light' : 'dark';
        }
        
        // Apply theme immediately for instant feedback
        html.setAttribute('data-bs-theme', newTheme);
        
        // Persist to server
        try {
            await fetch('/users/theme', {
                method: 'POST',
                body: JSON.stringify({ theme: newTheme }),
                headers: { 'Content-Type': 'application/json' }
            });
        } catch (err) {
            console.error('Failed to update theme:', err);
            // Revert if failed
            html.setAttribute('data-bs-theme', currentTheme);
            return;
        }
        
        // Update for next click
        updateIcon(newTheme);
        // Also update the currentTheme reference
        currentTheme = newTheme;
    });

    function updateIcon(theme) {
        if (!icon) return;
        if (theme === 'dark') {
            icon.className = 'fas fa-sun';
        } else if (theme === 'light') {
            icon.className = 'fas fa-moon';
        } else {
            // auto - show system preference
            const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            icon.className = systemPrefersDark ? 'fas fa-sun' : 'fas fa-moon';
        }
    }

    // Listen for system theme changes (for auto mode)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        const currentTheme = html.getAttribute('data-bs-theme');
        if (currentTheme === 'auto') {
            updateIcon(currentTheme);
        }
    });
});
