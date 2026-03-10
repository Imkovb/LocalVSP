package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var validDomainRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

func autoStartConfigPath() string {
	if path := os.Getenv("LOCALVSP_AUTOSTART_FILE"); path != "" {
		return path
	}
	return autoStartFile
}

func configEnvPath() string {
	if path := os.Getenv("LOCALVSP_ENV_FILE"); path != "" {
		return path
	}
	return "/opt/localvsp/.env"
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) //nolint:errcheck

	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close() //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func validateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return fmt.Errorf("invalid domain")
	}
	if !validDomainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain")
	}
	return nil
}

func parseEnvLines(data string) ([]string, map[string]string) {
	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	values := map[string]string{}
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		values[parts[0]] = parts[1]
	}
	return lines, values
}

func mergeEnv(existing string, updates map[string]string) string {
	lines, values := parseEnvLines(existing)
	used := map[string]bool{}
	merged := make([]string, 0, len(lines)+len(updates))

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			if strings.TrimSpace(line) != "" {
				merged = append(merged, line)
			}
			continue
		}
		key := parts[0]
		if newValue, ok := updates[key]; ok {
			merged = append(merged, key+"="+newValue)
			used[key] = true
		} else {
			merged = append(merged, key+"="+values[key])
		}
	}

	var extraKeys []string
	for key := range updates {
		if !used[key] {
			extraKeys = append(extraKeys, key)
		}
	}
	sort.Strings(extraKeys)
	for _, key := range extraKeys {
		merged = append(merged, key+"="+updates[key])
	}

	return strings.Join(merged, "\n") + "\n"
}