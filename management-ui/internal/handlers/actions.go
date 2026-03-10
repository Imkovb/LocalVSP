package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/localvsp/management-ui/internal/docker"
)

func InfraActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	service := strings.TrimSpace(r.FormValue("service"))
	action := strings.TrimSpace(r.FormValue("action"))
	if service == "" || action == "" {
		respondError(w, "service and action are required")
		return
	}

	if err := docker.InfraAction(service, action); err != nil {
		respondError(w, err.Error())
		return
	}

	InfraTableHandler(w, r)
}

func LocalHtmlActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	action := strings.TrimSpace(r.FormValue("action"))
	confirmedName := strings.TrimSpace(r.FormValue("confirmed_name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}

	switch action {
	case "deploy":
		if _, err := docker.DeployHtmlSite(name); err != nil {
			respondError(w, err.Error())
			return
		}
	case "stop":
		if err := docker.StopHtmlSite(name); err != nil {
			respondError(w, err.Error())
			return
		}
	case "delete":
		if confirmedName != name {
			respondError(w, "deletion confirmation did not match the site name")
			return
		}
		if err := docker.DeleteHtmlSite(name); err != nil {
			respondError(w, err.Error())
			return
		}
	default:
		respondError(w, "unknown action: "+action)
		return
	}

	LocalHtmlSitesHandler(w, r)
}

func LocalDockerActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	action := strings.TrimSpace(r.FormValue("action"))
	confirmedName := strings.TrimSpace(r.FormValue("confirmed_name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}

	switch action {
	case "deploy":
		startBuildJob(name)
		w.Header().Set("HX-Trigger-After-Swap", fmt.Sprintf(`{"openBuildPanel": %q}`, name))
		LocalDockerProjectsHandler(w, r)
		return
	case "stop":
		if err := docker.StopDockerProject(name); err != nil {
			respondError(w, err.Error())
			return
		}
	case "delete":
		if confirmedName != name {
			respondError(w, "deletion confirmation did not match the project name")
			return
		}
		if err := docker.DeleteDockerProject(name); err != nil {
			respondError(w, err.Error())
			return
		}
	default:
		respondError(w, "unknown action: "+action)
		return
	}

	LocalDockerProjectsHandler(w, r)
}

func LocalHtmlAutoStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}
	if err := docker.ToggleHtmlAutoStart(name, r.FormValue("enable") == "true"); err != nil {
		respondError(w, err.Error())
		return
	}

	LocalHtmlSitesHandler(w, r)
}

func LocalDockerAutoStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}
	if err := docker.ToggleDockerAutoStart(name, r.FormValue("enable") == "true"); err != nil {
		respondError(w, err.Error())
		return
	}

	LocalDockerProjectsHandler(w, r)
}

func LocalHtmlSubdomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}
	if err := docker.SetHtmlSubdomain(name, r.FormValue("subdomain")); err != nil {
		respondError(w, err.Error())
		return
	}

	LocalHtmlSitesHandler(w, r)
}

func LocalDockerSubdomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	name := sanitizeName(r.FormValue("name"))
	if name == "" || name == "." || name == "/" {
		respondError(w, "name is required")
		return
	}
	if err := docker.SetDockerSubdomain(name, r.FormValue("subdomain")); err != nil {
		respondError(w, err.Error())
		return
	}

	LocalDockerProjectsHandler(w, r)
}

func SaveSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		respondError(w, "invalid form data")
		return
	}

	prevCfg := docker.GetCurrentConfig()
	cfg := docker.Config{
		CFToken: strings.TrimSpace(r.FormValue("cf_token")),
		Domain:  strings.TrimSpace(r.FormValue("domain")),
	}
	if cfg.CFToken == "" && r.FormValue("clear_cf_token") != "true" {
		cfg.CFToken = prevCfg.CFToken
	}
	if cfg.Domain == "" && r.FormValue("clear_domain") != "true" {
		cfg.Domain = prevCfg.Domain
	}

	if err := docker.SaveConfig(cfg); err != nil {
		respondError(w, "Failed to save settings: "+err.Error())
		return
	}

	go func() {
		if err := docker.ApplyPlatformRoutingConfig(); err != nil {
			fmt.Printf("platform routing apply failed: %v\n", err)
		}
	}()

	msg := "Settings saved. Traefik routing and deployed app exposure are being refreshed."
	if cfg.CFToken != "" {
		msg = "Settings saved. Cloudflare Tunnel, Traefik routing, and deployed app exposure are being refreshed."
		if prevCfg.CFToken != "" && prevCfg.CFToken == cfg.CFToken {
			msg = "Settings saved. Cloudflare Tunnel and Traefik routing are being reapplied."
		}
	}
	respondSuccess(w, msg)
}