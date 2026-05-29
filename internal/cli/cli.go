package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/funkykay/summarize/internal/buildinfo"
	"github.com/funkykay/summarize/internal/initconfig"
	"github.com/funkykay/summarize/internal/summary"
	"github.com/funkykay/summarize/internal/update"
)

type App struct {
	stdout io.Writer
	stderr io.Writer
}

type globalOptions struct {
	baseDir string
	profile string
	dryRun  bool
}

func New(stdout io.Writer, stderr io.Writer) App {
	return App{stdout: stdout, stderr: stderr}
}

func (a App) Run(args []string) int {
	options, command, commandArgs, err := parseGlobalOptions(args)
	if err != nil {
		fmt.Fprintf(a.stderr, "Error: %s\n", err)
		return 1
	}

	if command == "" {
		if err := summary.Create(options.profile, options.baseDir, a.stdout, options.dryRun); err != nil {
			fmt.Fprintf(a.stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	}

	switch command {
	case "init":
		if err := a.runInit(options.baseDir, commandArgs); err != nil {
			fmt.Fprintf(a.stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "update":
		if err := a.runUpdate(commandArgs); err != nil {
			fmt.Fprintf(a.stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "version":
		if len(commandArgs) > 0 {
			fmt.Fprintf(a.stderr, "Error: unexpected argument for version: %s\n", commandArgs[0])
			return 1
		}
		fmt.Fprint(a.stdout, buildinfo.Version)
		return 0
	default:
		fmt.Fprintf(a.stderr, "Error: unknown command: %s\n", command)
		return 1
	}
}

func (a App) runInit(defaultBaseDir string, args []string) error {
	baseDir := defaultBaseDir

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--base-dir":
			if i+1 >= len(args) {
				return errors.New("--base-dir requires a value")
			}
			baseDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--base-dir="):
			baseDir = strings.TrimPrefix(arg, "--base-dir=")
		default:
			return fmt.Errorf("unexpected argument for init: %s", arg)
		}
	}

	resolvedBaseDir, err := resolveExistingDirectory(baseDir)
	if err != nil {
		return err
	}

	configPath := filepath.Join(resolvedBaseDir, "summarize.json")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("Configuration file already exists: %s", configPath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	configData := initconfig.BuildInitialConfig(resolvedBaseDir)
	encoded, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')

	if err := os.WriteFile(configPath, encoded, 0o644); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Created %s\n", configPath)
	return nil
}

func (a App) runUpdate(args []string) error {
	repo := ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--repo":
			if i+1 >= len(args) {
				return errors.New("--repo requires a value")
			}
			repo = args[i+1]
			i++
		case strings.HasPrefix(arg, "--repo="):
			repo = strings.TrimPrefix(arg, "--repo=")
		default:
			return fmt.Errorf("unexpected argument for update: %s", arg)
		}
	}

	return update.Binary(repo, a.stdout)
}

func parseGlobalOptions(args []string) (globalOptions, string, []string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return globalOptions{}, "", nil, err
	}

	options := globalOptions{baseDir: cwd}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--base-dir":
			if i+1 >= len(args) {
				return globalOptions{}, "", nil, errors.New("--base-dir requires a value")
			}
			options.baseDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--base-dir="):
			options.baseDir = strings.TrimPrefix(arg, "--base-dir=")
		case arg == "-p" || arg == "--profile":
			if i+1 >= len(args) {
				return globalOptions{}, "", nil, fmt.Errorf("%s requires a value", arg)
			}
			options.profile = args[i+1]
			i++
		case strings.HasPrefix(arg, "--profile="):
			options.profile = strings.TrimPrefix(arg, "--profile=")
		case arg == "--dry-run":
			options.dryRun = true
		case strings.HasPrefix(arg, "-"):
			return globalOptions{}, "", nil, fmt.Errorf("unknown option: %s", arg)
		default:
			resolvedBaseDir, err := resolveExistingDirectory(options.baseDir)
			if err != nil {
				return globalOptions{}, "", nil, err
			}
			options.baseDir = resolvedBaseDir
			return options, arg, args[i+1:], nil
		}
	}

	resolvedBaseDir, err := resolveExistingDirectory(options.baseDir)
	if err != nil {
		return globalOptions{}, "", nil, err
	}
	options.baseDir = resolvedBaseDir

	return options, "", nil, nil
}

func resolveExistingDirectory(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("Directory does not exist: %s", path)
		}
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("Expected a directory: %s", path)
	}

	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return resolved, nil
}
