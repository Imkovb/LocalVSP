package handlers

import (
	"net/http"

	"github.com/localvsp/management-ui/internal/docker"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	lang := requestLang(w, r)
	systemInfo, _ := docker.GetSystemInfo()

	renderTemplate(w, "dashboard.html", tplData(lang, map[string]interface{}{
		"SystemInfo": systemInfo,
		"Host":       hostFromRequest(r),
	}))
}

func LogsPageHandler(w http.ResponseWriter, r *http.Request) {
	appName := sanitizeName(r.URL.Query().Get("app"))
	if appName == "" || appName == "." || appName == "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	renderTemplate(w, "logs.html", tplData(requestLang(w, r), map[string]interface{}{
		"AppName": appName,
	}))
}

func HelpHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "help.html", tplData(requestLang(w, r), map[string]interface{}{
		"Host": hostFromRequest(r),
	}))
}

func SettingsPageHandler(w http.ResponseWriter, r *http.Request) {
	cfg := docker.GetCurrentConfig()
	renderTemplate(w, "settings.html", tplData(requestLang(w, r), map[string]interface{}{
		"CFToken": cfg.CFToken,
		"Domain":  cfg.Domain,
		"Host":    hostFromRequest(r),
	}))
}