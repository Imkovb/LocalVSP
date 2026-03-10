package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/localvsp/management-ui/internal/i18n"
	"github.com/localvsp/management-ui/internal/view"
)

func tplData(lang string, extra map[string]interface{}) map[string]interface{} {
	d := map[string]interface{}{
		"T":        i18n.T(lang),
		"Lang":     lang,
		"Langs":    i18n.SupportedLangs,
		"LangName": i18n.LangName,
		"LangFlag": i18n.LangFlag,
	}
	for k, v := range extra {
		d[k] = v
	}
	return d
}

func renderTemplate(w http.ResponseWriter, name string, data any) {
	view.Render(w, name, data)
}

func respondError(w http.ResponseWriter, msg string) {
	escaped := template.HTMLEscapeString(msg)
	fmt.Fprintf(w, `<div hx-swap-oob="beforeend:#notification-stack">
		<div class="notify-toast notify-toast-error" data-notify-toast>
			<span class="notify-toast-accent" aria-hidden="true"></span>
			<div class="notify-toast-body">%s</div>
			<button type="button" class="notify-toast-close" data-notify-close aria-label="Close">&times;</button>
		</div>
	</div>`, escaped)
}

func respondSuccess(w http.ResponseWriter, msg string) {
	escaped := template.HTMLEscapeString(msg)
	fmt.Fprintf(w, `<div hx-swap-oob="beforeend:#notification-stack">
		<div class="notify-toast notify-toast-success" data-notify-toast>
			<span class="notify-toast-accent" aria-hidden="true"></span>
			<div class="notify-toast-body">%s</div>
			<button type="button" class="notify-toast-close" data-notify-close aria-label="Close">&times;</button>
		</div>
	</div>`, escaped)
}

func sanitizeName(name string) string {
	return filepath.Clean(filepath.Base(name))
}

func hostFromRequest(r *http.Request) string {
	host := r.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func requestLang(w http.ResponseWriter, r *http.Request) string {
	return i18n.Detect(w, r)
}