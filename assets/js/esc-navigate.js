import { enableEscToNavigate } from '/js/keyboard.js';

// A page opts in with an element carrying data-esc-navigate="<url>"
const target = document.querySelector('[data-esc-navigate]');
if (target) {
    enableEscToNavigate(target.dataset.escNavigate);
}
