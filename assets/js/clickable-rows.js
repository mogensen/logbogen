// Rows with a data-href navigate there when clicked anywhere, links and buttons inside keep their own behavior
document.addEventListener('click', function (e) {
    const row = e.target.closest('tr[data-href]');
    if (!row || e.target.closest('a, button')) {
        return;
    }
    // Selecting text in a row must not navigate
    if (window.getSelection().toString() !== '') {
        return;
    }
    if (e.metaKey || e.ctrlKey) {
        window.open(row.dataset.href, '_blank');
        return;
    }
    window.location.href = row.dataset.href;
});
