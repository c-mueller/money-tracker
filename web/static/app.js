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

    // Sortable tables
    document.querySelectorAll('table[data-sortable]').forEach(function(table) {
        var headers = table.querySelectorAll('th[data-sort-type]');
        headers.forEach(function(th, colIndex) {
            th.classList.add('sortable');
            th.addEventListener('click', function() {
                var tbody = table.querySelector('tbody');
                if (!tbody) return;

                var rows = Array.from(tbody.querySelectorAll('tr'));
                var sortType = th.dataset.sortType;
                var asc = !th.classList.contains('sort-asc');

                // Reset all headers in this table
                headers.forEach(function(h) {
                    h.classList.remove('sort-asc', 'sort-desc');
                });
                th.classList.add(asc ? 'sort-asc' : 'sort-desc');

                // Find column index relative to all th in the row
                var allThs = Array.from(th.parentElement.children);
                var ci = allThs.indexOf(th);

                rows.sort(function(a, b) {
                    var cellA = a.children[ci];
                    var cellB = b.children[ci];
                    if (!cellA || !cellB) return 0;

                    var valA = cellA.dataset.sortValue || cellA.textContent.trim();
                    var valB = cellB.dataset.sortValue || cellB.textContent.trim();

                    var cmp = 0;
                    if (sortType === 'number') {
                        cmp = (parseFloat(valA) || 0) - (parseFloat(valB) || 0);
                    } else if (sortType === 'date') {
                        cmp = valA.localeCompare(valB);
                    } else {
                        cmp = valA.localeCompare(valB, undefined, { sensitivity: 'base' });
                    }

                    return asc ? cmp : -cmp;
                });

                rows.forEach(function(row) {
                    tbody.appendChild(row);
                });
            });
        });
    });
});
