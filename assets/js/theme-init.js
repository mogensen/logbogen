// Runs in the head, before the page paints, so a dark system theme does not flash light first
if (window.matchMedia('(prefers-color-scheme: dark)').matches && !document.documentElement.getAttribute('data-bs-theme')) {
    document.documentElement.setAttribute('data-bs-theme', 'dark');
}
