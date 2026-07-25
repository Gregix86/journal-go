document.querySelectorAll('form.confirm-submit').forEach(function (form) {
  form.addEventListener('submit', function (e) {
    const message = form.dataset.confirm || 'Confirmer cette action ?';
    if (!confirm(message)) {
      e.preventDefault();
    }
  });
});
