// Replaces inline onclick handlers: data-print opens the print dialog, data-confirm asks before a click goes on
document.addEventListener('click', (event) => {
    const target = event.target instanceof Element ? event.target : null;
    if (!target) return;

    const confirmEl = target.closest('[data-confirm]');
    if (confirmEl && !window.confirm(confirmEl.dataset.confirm)) {
        event.preventDefault();
        return;
    }

    if (target.closest('[data-print]')) {
        window.print();
    }
});
