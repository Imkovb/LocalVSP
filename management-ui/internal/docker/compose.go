package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func ComposeUp(dir string) error {
	if err := requireComposeFile(dir); err != nil {
		return err
	}
	return runCompose(dir, "up", "--build", "--detach", "--remove-orphans")
}

func ComposeUpStream(dir string, send func(string)) error {
	if err := requireComposeFile(dir); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "--build", "--detach", "--remove-orphans")
	cmd.Dir = dir
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker compose: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	reader := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			send(scanner.Text())
		}
	}
	go reader(stdoutPipe)
	go reader(stderrPipe)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker compose timed out after 10 minutes")
		}
		return fmt.Errorf("docker compose up failed: %v", err)
	}
	return nil
}

func ComposeStop(dir string) error {
	if err := requireComposeFile(dir); err != nil {
		return err
	}
	return runCompose(dir, "stop")
}

func ComposeRestart(dir string) error {
	if err := requireComposeFile(dir); err != nil {
		return err
	}
	return runCompose(dir, "restart")
}

func ComposeDown(dir string) error {
	if err := requireComposeFile(dir); err != nil {
		return err
	}
	return runCompose(dir, "down")
}

func ComposeLogs(dir, lines string) (string, error) {
	if err := requireComposeFile(dir); err != nil {
		return "", err
	}
	output, err := runDockerCmd(defaultTimeout, dir, "compose", "logs", "--no-color", "--tail="+lines)
	if err != nil {
		return "", fmt.Errorf("%s", string(output))
	}
	return string(output), nil
}

var composeFileCandidates = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

func findComposeFilePath(dir string) (string, bool) {
	for _, name := range composeFileCandidates {
		path := filepath.Join(dir, name)
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			return path, true
		}
	}
	return "", false
}

func requireComposeFile(dir string) error {
	if _, ok := findComposeFilePath(dir); !ok {
		return fmt.Errorf("no compose file found in %s (expected one of: docker-compose.yml, docker-compose.yaml, compose.yml, compose.yaml)", dir)
	}
	return nil
}

func runCompose(dir string, args ...string) error {
	output, err := runDockerCmd(buildTimeout, dir, append([]string{"compose"}, args...)...)
	if err != nil {
		return fmt.Errorf("docker compose %s failed: %s", args[0], string(output))
	}
	return nil
}

func firstServiceFromComposeFile(dir string) string {
	composePath, ok := findComposeFilePath(dir)
	if !ok {
		return ""
	}

	b, err := os.ReadFile(composePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(b), "\n")
	servicesIndent := -1
	serviceLevelIndent := -1

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if servicesIndent == -1 {
			if trimmed == "services:" {
				servicesIndent = indent
			}
			continue
		}

		if indent <= servicesIndent {
			break
		}
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") || !strings.Contains(trimmed, ":") {
			continue
		}
		if serviceLevelIndent == -1 {
			serviceLevelIndent = indent
		}
		if indent != serviceLevelIndent {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.Trim(strings.TrimSpace(parts[0]), `"'`)
		if name != "" {
			return name
		}
	}

	return ""
}

func getFirstComposeService(dir string) string {
	if s := firstServiceFromComposeFile(dir); s != "" {
		return s
	}

	out, err := runDockerCmd(defaultTimeout, dir, "compose", "config", "--services")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return ""
}

func getComposeServicesStatus(dir string) ([]Service, string) {
	out, err := runDockerCmd(defaultTimeout, dir, "compose", "ps", "--format", "json")
	if err != nil {
		return nil, "unknown"
	}

	type composePS struct {
		Name       string `json:"Name"`
		Image      string `json:"Image"`
		State      string `json:"State"`
		Publishers []struct {
			PublishedPort int `json:"PublishedPort"`
			TargetPort    int `json:"TargetPort"`
		} `json:"Publishers"`
	}

	var services []Service
	runningCount := 0
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var ps composePS
		if err := json.Unmarshal([]byte(line), &ps); err != nil {
			continue
		}
		ports := ""
		for _, p := range ps.Publishers {
			if p.PublishedPort > 0 {
				ports += strconv.Itoa(p.PublishedPort) + "->" + strconv.Itoa(p.TargetPort) + " "
			}
		}
		services = append(services, Service{Name: ps.Name, Image: ps.Image, Status: ps.State, Ports: strings.TrimSpace(ports)})
		if strings.ToLower(ps.State) == "running" {
			runningCount++
		}
	}

	aggregateStatus := "stopped"
	if len(services) == 0 {
		aggregateStatus = "unknown"
	} else if runningCount == len(services) {
		aggregateStatus = "running"
	} else if runningCount > 0 {
		aggregateStatus = "partial"
	}

	return services, aggregateStatus
}