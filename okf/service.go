package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const launchdLabel = "com.samcharles93.okf-mcp"

const systemdUnitName = "okf-mcp.service"

// ServiceConfig describes an `okf mcp` invocation to register as a
// background OS service.
type ServiceConfig struct {
	BundleRoot string
	Addr       string
	BinPath    string
}

// SystemdUnit renders a systemd user unit file that runs `okf mcp` as a
// background service on Linux.
func SystemdUnit(cfg ServiceConfig) string {
	return fmt.Sprintf(`[Unit]
Description=OKF Knowledge Bundle MCP server
After=network.target

[Service]
Type=simple
ExecStart=%s mcp --bundle %s --addr %s
Restart=on-failure
RestartSec=2

[Install]
WantedBy=default.target
`, cfg.BinPath, cfg.BundleRoot, cfg.Addr)
}

// LaunchdPlist renders a launchd LaunchAgent property list that runs
// `okf mcp` as a background service on macOS.
func LaunchdPlist(cfg ServiceConfig) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>mcp</string>
		<string>--bundle</string>
		<string>%s</string>
		<string>--addr</string>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`, launchdLabel, cfg.BinPath, cfg.BundleRoot, cfg.Addr)
}

// ServiceStatus reports whether `okf mcp` is registered as a background
// service, and its best-effort live state.
type ServiceStatus struct {
	Installed bool
	Path      string
	State     string
}

// ServiceInstaller manages OS-level background service registration for
// `okf mcp`. GOOS, HomeDir, and Run are overridable so tests never touch
// the real service manager or home directory; production callers should
// set GOOS: runtime.GOOS, HomeDir: result of os.UserHomeDir(), and Run
// wrapping exec.Command.
type ServiceInstaller struct {
	GOOS    string
	HomeDir string
	Run     func(name string, args ...string) (string, error)
}

// unitPath resolves where the service definition file lives for the
// installer's GOOS, or an error if the platform isn't supported.
func (i ServiceInstaller) unitPath() (string, error) {
	switch i.GOOS {
	case "linux":
		return filepath.Join(i.HomeDir, ".config", "systemd", "user", systemdUnitName), nil
	case "darwin":
		return filepath.Join(i.HomeDir, "Library", "LaunchAgents", launchdLabel+".plist"), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s (service installation supports linux and darwin only)", i.GOOS)
	}
}

// Install writes the service definition file for cfg. If enable is true,
// it also registers and starts the service with the platform's service
// manager (systemctl --user daemon-reload && enable --now on Linux,
// launchctl load -w on macOS).
func (i ServiceInstaller) Install(cfg ServiceConfig, enable bool) (string, error) {
	path, err := i.unitPath()
	if err != nil {
		return "", err
	}

	var content string
	switch i.GOOS {
	case "linux":
		content = SystemdUnit(cfg)
	case "darwin":
		content = LaunchdPlist(cfg)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}

	if !enable {
		return path, nil
	}

	switch i.GOOS {
	case "linux":
		if _, err := i.Run("systemctl", "--user", "daemon-reload"); err != nil {
			return path, err
		}
		if _, err := i.Run("systemctl", "--user", "enable", "--now", systemdUnitName); err != nil {
			return path, err
		}
	case "darwin":
		if _, err := i.Run("launchctl", "load", "-w", path); err != nil {
			return path, err
		}
	}

	return path, nil
}

// Uninstall removes the service definition file, reporting whether one
// was found at all. If disable is true, it first attempts to stop the
// service via the platform's service manager — a failure there (e.g. the
// unit was never actually loaded) does not prevent the file from being
// removed.
func (i ServiceInstaller) Uninstall(disable bool) (bool, string, error) {
	path, err := i.unitPath()
	if err != nil {
		return false, "", err
	}

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return false, path, nil
	} else if statErr != nil {
		return false, "", statErr
	}

	if disable {
		switch i.GOOS {
		case "linux":
			_, _ = i.Run("systemctl", "--user", "disable", "--now", systemdUnitName)
		case "darwin":
			_, _ = i.Run("launchctl", "unload", path)
		}
	}

	if err := os.Remove(path); err != nil {
		return true, path, err
	}
	return true, path, nil
}

// Status reports whether `okf mcp` is registered as a service. If query
// is true and the service is installed, it also best-effort queries live
// state from the platform's service manager; a failed query degrades to
// a descriptive State rather than an error, since the file existing is
// the authoritative "installed" signal regardless of live query success.
func (i ServiceInstaller) Status(query bool) (ServiceStatus, error) {
	path, err := i.unitPath()
	if err != nil {
		return ServiceStatus{}, err
	}

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return ServiceStatus{Installed: false, Path: path}, nil
	} else if statErr != nil {
		return ServiceStatus{}, statErr
	}

	status := ServiceStatus{Installed: true, Path: path}
	if !query {
		return status, nil
	}

	var out string
	var runErr error
	switch i.GOOS {
	case "linux":
		out, runErr = i.Run("systemctl", "--user", "is-active", systemdUnitName)
	case "darwin":
		out, runErr = i.Run("launchctl", "list", launchdLabel)
	}
	if runErr != nil {
		status.State = fmt.Sprintf("unknown (live query failed: %v)", runErr)
	} else {
		status.State = strings.TrimSpace(out)
	}
	return status, nil
}
