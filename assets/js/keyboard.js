// Keyboard navigation utilities

/**
 * Enables ESC key navigation to a given URL.
 * @param {string} backUrl - The URL to navigate to when ESC is pressed.
 */
export function enableEscToNavigate(backUrl) {
  document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
      window.location.href = backUrl;
    }
  });
} 