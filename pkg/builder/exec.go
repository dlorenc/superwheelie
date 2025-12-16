package builder

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// DefaultTimeout is the default timeout for build commands.
const DefaultTimeout = 4 * time.Hour

// ExecResult contains the result of executing a command.
type ExecResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Error    error
}

// Exec runs a command and returns the result.
func Exec(ctx context.Context, dir string, env []string, name string, args ...string) *ExecResult {
	start := time.Now()
	result := &ExecResult{
		Command: fmt.Sprintf("%s %v", name, args),
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// ExecWithTimeout runs a command with a timeout.
func ExecWithTimeout(timeout time.Duration, dir string, env []string, name string, args ...string) *ExecResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return Exec(ctx, dir, env, name, args...)
}

// ExecSimple runs a command and returns stdout or an error.
func ExecSimple(dir string, name string, args ...string) (string, error) {
	result := ExecWithTimeout(DefaultTimeout, dir, nil, name, args...)
	if result.Error != nil {
		return "", fmt.Errorf("%s: %w\nstdout: %s\nstderr: %s",
			result.Command, result.Error, result.Stdout, result.Stderr)
	}
	return result.Stdout, nil
}

// CombinedOutput returns combined stdout and stderr from an ExecResult.
func (r *ExecResult) CombinedOutput() string {
	if r.Stderr == "" {
		return r.Stdout
	}
	if r.Stdout == "" {
		return r.Stderr
	}
	return r.Stdout + "\n" + r.Stderr
}

// Success returns true if the command succeeded.
func (r *ExecResult) Success() bool {
	return r.Error == nil && r.ExitCode == 0
}
