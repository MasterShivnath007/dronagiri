// Package runner executes external commands with a predictable environment.
package runner

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Runner runs commands the way a terminal user would, with a PATH that
// includes the usual macOS install locations.
type Runner struct {
	Env []string
}

// New builds a runner whose PATH always contains the Homebrew and user bin
// directories, even when dronagiri is launched from a minimal environment.
func New() *Runner {
	path := os.Getenv("PATH")
	home, _ := os.UserHomeDir()
	extra := []string{
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
		home + "/.local/bin",
		home + "/.cargo/bin",
		home + "/go/bin",
	}
	for _, dir := range extra {
		if dir == "" || strings.Contains(path, dir) {
			continue
		}
		if path == "" {
			path = dir
		} else {
			path += ":" + dir
		}
	}

	env := os.Environ()
	env = append(env,
		"PATH="+path,
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"HOMEBREW_NO_ANALYTICS=1",
		"HOMEBREW_NO_ENV_HINTS=1",
		"NONINTERACTIVE=1",
		"NO_COLOR=1",
	)
	return &Runner{Env: env}
}

// Output runs a command and returns its stdout. Stderr is captured and
// included in the error when the command fails.
func (r *Runner) Output(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = r.Env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil

	err := cmd.Run()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return stdout.String(), errors.New(detail)
	}
	return stdout.String(), nil
}

// Stream runs a command and returns its combined output. Used for actions
// such as uninstalling, where seeing the manager's own output matters.
func (r *Runner) Stream(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = r.Env
	cmd.Stdin = nil
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Available reports whether an executable is on PATH.
func (r *Runner) Available(name string) bool {
	_, err := r.Lookup(name)
	return err == nil
}

// Lookup finds an executable using the runner's PATH.
func (r *Runner) Lookup(name string) (string, error) {
	if strings.Contains(name, "/") {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
		return "", errors.New("not found")
	}
	for _, entry := range r.Env {
		if !strings.HasPrefix(entry, "PATH=") {
			continue
		}
		for _, dir := range strings.Split(strings.TrimPrefix(entry, "PATH="), ":") {
			candidate := dir + "/" + name
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", errors.New("not found")
}

// DuSizes returns the on-disk size of each path in a single `du` call.
// Missing paths are simply absent from the result.
func (r *Runner) DuSizes(paths []string) map[string]int64 {
	sizes := make(map[string]int64, len(paths))
	if len(paths) == 0 {
		return sizes
	}

	// du needs reasonably sized argument lists; chunk to stay well under ARG_MAX.
	const chunk = 200
	for start := 0; start < len(paths); start += chunk {
		end := start + chunk
		if end > len(paths) {
			end = len(paths)
		}
		args := append([]string{"-s", "-k"}, paths[start:end]...)
		out, err := r.Output("du", args...)
		if err != nil && out == "" {
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(out))
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			idx := strings.IndexByte(line, '\t')
			if idx < 0 {
				continue
			}
			kb, convErr := strconv.ParseInt(strings.TrimSpace(line[:idx]), 10, 64)
			if convErr != nil {
				continue
			}
			sizes[line[idx+1:]] = kb * 1024
		}
	}
	return sizes
}
