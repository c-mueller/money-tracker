# 005 — Household Edit Tab, Icons, Self-Hosted Assets, Category Edit

## Context

Several UX improvements: Households could not be edited via the web UI. Categories and households should get Material Icons for visual identification. All external dependencies (Bootstrap, Icons) should be self-hosted. Categories should be editable (name + icon). Inline-created categories get a generic default icon.

## Implemented Changes

### Self-Hosted Assets
- Bootstrap 5.3.3 CSS + JS bundle in `web/static/vendor/`
- Google Material Symbols Outlined (WOFF2 + CSS) in `web/static/vendor/material-icons/`
- Layout.html: CDN links replaced with local paths

### Icon Field (Schema + Domain)
- `ent/schema/category.go`: New field `icon` (optional, MaxLen 50, default "category")
- `ent/schema/household.go`: New field `icon` (optional, MaxLen 50, default "home")
- Domain types, repository converters, create/update in repos updated
- Service signatures extended: `Create(ctx, ..., icon)` / `Update(ctx, ..., icon)`
- All callers (API handlers, web handlers) updated

### Icon Picker
- `web/static/icons.json`: Curated list of 50 Material Symbols
- `web/templates/partials/icon-picker.html`: Reusable partial with grid + JS
- Template renderer: `Icons []string` loaded, `dict` template function added

### Household Settings Tab
- New "Settings" tab in `household/tabs.html`
- `web/templates/household/settings.html`: Name, currency, icon picker, delete
- Routes: `GET/POST /households/:id/settings`
- Delete button moved from header to settings

### Category Edit
- `web/templates/category/form.html`: Edit form with name + icon picker
- `web/templates/category/list.html`: Icons next to names, clickable to edit
- Routes: `GET /households/:id/categories/:categoryId/edit`, `POST /households/:id/categories/:categoryId`
- `CategoryService.GetByID()` added

### Icons in Views
- Dashboard: Household icon next to name
- Household header (tabs.html): Icon next to name
- Category list: Icon next to name
