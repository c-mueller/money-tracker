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
                        alert('Error: ' + resp.statusText);
                    }
                })
                .catch(function(err) {
                    alert('Error: ' + err.message);
                });
        });
    });

    // Copy-to-clipboard buttons
    document.querySelectorAll('[data-copy]').forEach(function(btn) {
        btn.addEventListener('click', function() {
            var target = document.querySelector(this.dataset.copy);
            if (target) {
                navigator.clipboard.writeText(target.textContent.trim()).then(function() {
                    btn.textContent = 'Copied!';
                    setTimeout(function() { btn.textContent = 'Copy'; }, 2000);
                });
            }
        });
    });
});
