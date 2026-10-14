// Copies the error details of the error page, for a bug report
document.addEventListener('DOMContentLoaded', function () {
    const button = document.getElementById('copy-error-info');
    if (!button) {
        return;
    }
    button.addEventListener('click', function () {
        const heading = document.querySelector('.error-body h1').textContent;
        const code = document.querySelector('.error-code').textContent;
        const lines = [code + ': ' + heading];
        document.querySelectorAll('.error-details dl > div').forEach(function (row) {
            lines.push(row.querySelector('dt').textContent + ': ' + row.querySelector('dd').textContent);
        });
        if (navigator.clipboard) {
            navigator.clipboard.writeText(lines.join('\n'));
        }
        button.textContent = 'Kopieret ✓';
        setTimeout(function () { button.textContent = button.dataset.label; }, 1500);
    });
});
