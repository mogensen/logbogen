// Shows how much of the limit a text field has used. The browser stops typing at the field's maxlength.
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-counter]').forEach(function (field) {
        const counter = document.getElementById(field.dataset.counter);
        if (!counter) {
            return;
        }

        const update = function () {
            counter.textContent = field.value.length;
            counter.parentElement.classList.toggle('act-edit-count--near', field.value.length >= field.maxLength * 0.9);
        };
        field.addEventListener('input', update);
        update();
    });
});
