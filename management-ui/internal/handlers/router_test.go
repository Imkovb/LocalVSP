package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/localvsp/management-ui/internal/docker"
)

func TestNewMuxReturns404ForUnknownRoute(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rr := httptest.NewRecorder()

	NewMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSaveSettingsHandlerRejectsInvalidDomain(t *testing.T) {
	t.Parallel()

	t.Setenv("LOCALVSP_ENV_FILE", t.TempDir()+"/.env")
	form := strings.NewReader("cf_token=&domain=bad/domain")
	req := httptest.NewRequest(http.MethodPost, "/settings/save", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SaveSettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 error fragment, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to save settings") {
		t.Fatalf("expected error response, got %q", rr.Body.String())
	}
	if cfg := docker.GetCurrentConfig(); cfg.Domain != "" || cfg.CFToken != "" {
		t.Fatalf("expected config not to be written, got %+v", cfg)
	}
}

func TestSaveSettingsHandlerPreservesExistingValuesOnBlankSubmit(t *testing.T) {
	t.Parallel()

	envPath := t.TempDir() + "/.env"
	t.Setenv("LOCALVSP_ENV_FILE", envPath)
	if err := docker.SaveConfig(docker.Config{CFToken: "token-123", Domain: "beckhoven.nl"}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	form := strings.NewReader("cf_token=&domain=")
	req := httptest.NewRequest(http.MethodPost, "/settings/save", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SaveSettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", rr.Code)
	}
	if cfg := docker.GetCurrentConfig(); cfg.Domain != "beckhoven.nl" || cfg.CFToken != "token-123" {
		t.Fatalf("expected existing config to be preserved, got %+v", cfg)
	}
}

func TestSaveSettingsHandlerAllowsExplicitClears(t *testing.T) {
	t.Parallel()

	envPath := t.TempDir() + "/.env"
	t.Setenv("LOCALVSP_ENV_FILE", envPath)
	if err := docker.SaveConfig(docker.Config{CFToken: "token-123", Domain: "beckhoven.nl"}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	form := strings.NewReader("cf_token=&domain=&clear_cf_token=true&clear_domain=true")
	req := httptest.NewRequest(http.MethodPost, "/settings/save", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	SaveSettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", rr.Code)
	}
	if cfg := docker.GetCurrentConfig(); cfg.Domain != "" || cfg.CFToken != "" {
		t.Fatalf("expected config to be cleared, got %+v", cfg)
	}
}