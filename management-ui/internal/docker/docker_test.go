package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstServiceFromComposeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	compose := `services:
  web:
    image: nginx:alpine
  worker:
    image: busybox
`
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0644); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	if got := firstServiceFromComposeFile(dir); got != "web" {
		t.Fatalf("expected first service web, got %q", got)
	}
}

func TestMergeEnvPreservesUnknownKeys(t *testing.T) {
	t.Parallel()

	existing := "KEEP_ME=1\nVSP_DOMAIN=old.example\n"
	merged := mergeEnv(existing, map[string]string{
		"CLOUDFLARE_TUNNEL_TOKEN": "token-123",
		"VSP_DOMAIN":             "new.example",
	})

	if !strings.Contains(merged, "KEEP_ME=1") {
		t.Fatalf("expected unknown key to be preserved, got %q", merged)
	}
	if !strings.Contains(merged, "VSP_DOMAIN=new.example") {
		t.Fatalf("expected domain to be updated, got %q", merged)
	}
	if !strings.Contains(merged, "CLOUDFLARE_TUNNEL_TOKEN=token-123") {
		t.Fatalf("expected token to be written, got %q", merged)
	}
}

func TestSaveConfigPreservesExistingKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("KEEP_ME=1\nVSP_DOMAIN=old.local\n"), 0600); err != nil {
		t.Fatalf("seed env file: %v", err)
	}
	t.Setenv("LOCALVSP_ENV_FILE", envPath)

	if err := SaveConfig(Config{CFToken: "abc", Domain: "example.com"}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "KEEP_ME=1") {
		t.Fatalf("expected unrelated key to be preserved, got %q", text)
	}
	if !strings.Contains(text, "VSP_DOMAIN=example.com") {
		t.Fatalf("expected updated domain, got %q", text)
	}
	if !strings.Contains(text, "CLOUDFLARE_TUNNEL_TOKEN=abc") {
		t.Fatalf("expected token, got %q", text)
	}
	if cfg := GetCurrentConfig(); cfg.Domain != "example.com" || cfg.CFToken != "abc" {
		t.Fatalf("unexpected config readback: %+v", cfg)
	}
}

func TestValidateDomain(t *testing.T) {
	t.Parallel()

	for _, domain := range []string{"", "example.com", "sub.example.local"} {
		if err := validateDomain(domain); err != nil {
			t.Fatalf("expected valid domain %q, got %v", domain, err)
		}
	}

	for _, domain := range []string{" bad", ".example.com", "example..com", "example/com"} {
		if err := validateDomain(domain); err == nil {
			t.Fatalf("expected invalid domain %q", domain)
		}
	}
}

func TestTraefikRouterName(t *testing.T) {
	t.Parallel()

	got := traefikRouterName("vsp", "Hello.World/app")
	if got != "vsp-hello-world-app" {
		t.Fatalf("unexpected router name: %q", got)
	}
}

func TestShouldExposeThroughTraefik(t *testing.T) {
	t.Parallel()

	if !shouldExposeThroughTraefik("blog", "example.com") {
		t.Fatal("expected subdomain + domain to be exposed through traefik")
	}
	if shouldExposeThroughTraefik("blog", "") {
		t.Fatal("expected missing domain to disable traefik hostname exposure")
	}
	if shouldExposeThroughTraefik("", "example.com") {
		t.Fatal("expected missing subdomain to disable traefik hostname exposure")
	}
}

func TestGenerateDockerOverrideIncludesTraefikNetworkAndEntrypoint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	compose := `services:
  web:
    image: nginx:alpine
`
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0644); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	pc := ProjectConfig{
		Subdomain:  "blog",
		ExposePort: "8080",
		HostPort:   "8301",
	}
	if err := GenerateDockerOverride(dir, "My App", pc, "example.com"); err != nil {
		t.Fatalf("generate docker override: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "docker-compose.override.yml"))
	if err != nil {
		t.Fatalf("read override file: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		"traefik.enable=true",
		"traefik.docker.network=vsp-network",
		"traefik.http.routers.vsp-my-app.entrypoints=web",
		"traefik.http.routers.vsp-my-app.rule=Host(`blog.example.com`)",
		"traefik.http.services.vsp-my-app.loadbalancer.server.port=8080",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected override to contain %q, got %q", want, text)
		}
	}
}