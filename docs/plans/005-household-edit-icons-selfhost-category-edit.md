# 005 — Household-Edit Tab, Icons, Self-Hosted Assets, Category-Edit

## Kontext

Mehrere UX-Verbesserungen: Household lässt sich bisher nicht über die Web-UI editieren. Kategorien und Haushalte sollen Material Icons als visuelle Zuordnung bekommen. Alle externen Dependencies (Bootstrap, Icons) sollen lokal gehostet werden. Kategorien sollen editierbar sein (Name + Icon). Inline angelegte Kategorien bekommen ein generisches Default-Icon.

## Umgesetzte Änderungen

### Self-Hosted Assets
- Bootstrap 5.3.3 CSS + JS Bundle in `web/static/vendor/`
- Google Material Symbols Outlined (WOFF2 + CSS) in `web/static/vendor/material-icons/`
- Layout.html: CDN-Links durch lokale Pfade ersetzt

### Icon-Feld (Schema + Domain)
- `ent/schema/category.go`: Neues Feld `icon` (Optional, MaxLen 50, Default "category")
- `ent/schema/household.go`: Neues Feld `icon` (Optional, MaxLen 50, Default "home")
- Domain-Typen, Repository-Converter, Create/Update in Repos angepasst
- Service-Signaturen erweitert: `Create(ctx, ..., icon)` / `Update(ctx, ..., icon)`
- Alle Aufrufer (API-Handler, Web-Handler) angepasst

### Icon-Picker
- `web/static/icons.json`: Kuratierte Liste von 50 Material Symbols
- `web/templates/partials/icon-picker.html`: Wiederverwendbares Partial mit Grid + JS
- Template-Renderer: `Icons []string` geladen, `dict` Template-Funktion hinzugefügt

### Household Settings Tab
- Neuer Tab "Settings" in `household/tabs.html`
- `web/templates/household/settings.html`: Name, Currency, Icon-Picker, Delete
- Routes: `GET/POST /households/:id/settings`
- Delete-Button von Header in Settings verschoben

### Category Edit
- `web/templates/category/form.html`: Edit-Formular mit Name + Icon-Picker
- `web/templates/category/list.html`: Icons neben Namen, klickbar zu Edit
- Routes: `GET /households/:id/categories/:categoryId/edit`, `POST /households/:id/categories/:categoryId`
- `CategoryService.GetByID()` hinzugefügt

### Icons in Views
- Dashboard: Household-Icon neben Name
- Household-Header (tabs.html): Icon neben Name
- Category-Liste: Icon neben Name
