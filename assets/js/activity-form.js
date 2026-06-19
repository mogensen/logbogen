document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('activity-form');
    if (!form) return;

    const typeInput = document.getElementById('activity-type-input');
    const typeGroup = document.getElementById('activity-type-group');
    const typeError = document.getElementById('type-error');

    if (form) {
        form.addEventListener('submit', function(e) {
            if (!typeInput.value) {
                e.preventDefault();
                typeError.style.display = 'block';
                return false;
            }
            return true;
        });
    }

    if (typeGroup) {
        typeGroup.addEventListener('click', function(e) {
            if (e.target.tagName === 'INPUT') {
                typeInput.value = e.target.value;
                typeError.style.display = 'none';
            }
        });
    }
}); 