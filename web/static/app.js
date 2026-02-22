document.addEventListener('DOMContentLoaded', function() {
    // Confirm + API delete buttons
    document.querySelectorAll('[data-confirm]').forEach(function(btn) {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            if (!confirm(this.dataset.confirm)) return;

            var action = this.dataset.action;
            var method = this.dataset.method || 'DELETE';
            var redirect = this.dataset.redirect;

            fetch(action, { method: method })
                .then(function(resp) {
                    if (resp.ok) {
                        if (redirect) {
                            window.location.href = redirect;
                        } else {
                            window.location.reload();
                        }
                    } else {
                        alert((window.i18n && window.i18n.error_prefix || 'Error: ') + resp.statusText);
                    }
                })
                .catch(function(err) {
                    alert((window.i18n && window.i18n.error_prefix || 'Error: ') + err.message);
                });
        });
    });

    // Copy-to-clipboard buttons
    document.querySelectorAll('[data-copy]').forEach(function(btn) {
        btn.addEventListener('click', function() {
            var target = document.querySelector(this.dataset.copy);
            if (target) {
                var copyLabel = window.i18n && window.i18n.copy || 'Copy';
                var copiedLabel = window.i18n && window.i18n.copied || 'Copied!';
                navigator.clipboard.writeText(target.textContent.trim()).then(function() {
                    btn.textContent = copiedLabel;
                    setTimeout(function() { btn.textContent = copyLabel; }, 2000);
                });
            }
        });
    });
});
