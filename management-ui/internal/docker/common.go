package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	buildTimeout   = 10 * time.Minute
	homeVSP        = "/vsp-home"
	traefikNetwork = "vsp-network"
	autoStartFile  = "/vsp-home/.localvsp/autostart.json"
)

var (
	validNameRegex      = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	validSubdomainRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

type App struct {
	Name       string
	Status     string
	Services   []Service
	Dir        string
	DeployedAt time.Time
	HostPort   string
}

type Service struct {
	Name   string
	Image  string
	Status string
	Ports  string
}

type SystemInfo struct {
	DockerVersion     string
	TotalContainers   int
	RunningContainers int
	DiskUsage         string
	LoadAvg           string
	MemoryUsage       string
}

type Config struct {
	CFToken string
	Domain  string
}

func runCmd(timeout time.Duration, dir string, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("command timed out after %v: %s %v", timeout, name, args)
	}
	return out, err
}

func runDockerCmd(timeout time.Duration, dir string, args ...string) ([]byte, error) {
	return runCmd(timeout, dir, "docker", args...)
}

func isValidName(name string) bool {
	return validNameRegex.MatchString(name)
}

func isValidSubdomain(subdomain string) bool {
	return subdomain == "" || validSubdomainRegex.MatchString(subdomain)
}

func hostHomeVSP() string {
	if v := os.Getenv("VSP_HOST_HOME"); v != "" {
		return v
	}
	return "/home/vsp"
}

func traefikRouterName(prefix, name string) string {
	replacer := strings.NewReplacer(" ", "-", ".", "-", "/", "-", "\\", "-", ":", "-", "@", "-")
	clean := strings.ToLower(replacer.Replace(strings.TrimSpace(name)))
	clean = strings.Trim(clean, "-")
	if clean == "" {
		clean = "app"
	}
	return prefix + "-" + clean
}

func shouldExposeThroughTraefik(subdomain, domain string) bool {
	return strings.TrimSpace(subdomain) != "" && strings.TrimSpace(domain) != ""
}