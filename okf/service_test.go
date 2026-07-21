package okf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServiceConfig() ServiceConfig {
	return ServiceConfig{
		BundleRoot: "/home/sam/project/.okf",
		Addr:       ":8080",
		BinPath:    "/usr/local/bin/okf",
	}
}

func TestSystemdUnitRendersExecStartWithBundleAndAddr(t *testing.T) {
	t.Parallel()

	unit := SystemdUnit(testServiceConfig())
	if !strings.Contains(unit, "[Unit]") || !strings.Contains(unit, "[Service]") || !strings.Contains(unit, "[Install]") {
		t.Fatalf("SystemdUnit() missing standard sections:\n%s", unit)
	}
	wantExec := "ExecStart=/usr/local/bin/okf mcp --bundle /home/sam/project/.okf --addr :8080"
	if !strings.Contains(unit, wantExec) {
		t.Errorf("SystemdUnit() = %q, want ExecStart %q", unit, wantExec)
	}
	if !strings.Contains(unit, "Restart=on-failure") {
		t.Errorf("SystemdUnit() missing Restart=on-failure:\n%s", unit)
	}
	if !strings.Contains(unit, "WantedBy=default.target") {
		t.Errorf("SystemdUnit() missing WantedBy=default.target:\n%s", unit)
	}
}

func TestLaunchdPlistRendersProgramArguments(t *testing.T) {
	t.Parallel()

	plist := LaunchdPlist(testServiceConfig())
	for _, want := range []string{
		"<key>Label</key>",
		launchdLabel,
		"<key>ProgramArguments</key>",
		"<string>/usr/local/bin/okf</string>",
		"<string>mcp</string>",
		"<string>--bundle</string>",
		"<string>/home/sam/project/.okf</string>",
		"<string>--addr</string>",
		"<string>:8080</string>",
		"<key>RunAtLoad</key>",
		"<key>KeepAlive</key>",
	} {
		if !strings.Contains(plist, want) {
			t.Errorf("LaunchdPlist() missing %q:\n%s", want, plist)
		}
	}
}

// fakeRunner records every invocation and returns canned (output, error)
// pairs in call order, so tests can assert exact command sequences without
// touching a real service manager.
type fakeRunner struct {
	calls   [][]string
	outputs []string
	errs    []error
	i       int
}

func (f *fakeRunner) run(name string, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	if f.i >= len(f.outputs) {
		return "", nil
	}
	out, err := f.outputs[f.i], f.errs[f.i]
	f.i++
	return out, err
}

func (f *fakeRunner) queue(output string, err error) {
	f.outputs = append(f.outputs, output)
	f.errs = append(f.errs, err)
}

func TestServiceInstallerInstallWritesSystemdUnitOnLinux(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	wantPath := filepath.Join(home, ".config", "systemd", "user", "okf-mcp.service")
	if path != wantPath {
		t.Errorf("Install() path = %q, want %q", path, wantPath)
	}
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("service file not written: %v", err)
	}
	if !strings.Contains(string(data), "ExecStart=/usr/local/bin/okf mcp") {
		t.Errorf("service file content = %q", data)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("Run() calls = %v, want 2 invocations (daemon-reload, enable --now)", runner.calls)
	}
	if runner.calls[0][0] != "systemctl" || runner.calls[0][len(runner.calls[0])-1] != "daemon-reload" {
		t.Errorf("first call = %v, want systemctl ... daemon-reload", runner.calls[0])
	}
	last := runner.calls[1]
	if last[0] != "systemctl" || !strings.Contains(strings.Join(last, " "), "enable --now okf-mcp.service") {
		t.Errorf("second call = %v, want systemctl --user enable --now okf-mcp.service", last)
	}
}

func TestServiceInstallerInstallWritesLaunchdPlistOnDarwin(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "darwin", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	wantPath := filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
	if path != wantPath {
		t.Errorf("Install() path = %q, want %q", path, wantPath)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("plist not written: %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("Run() calls = %v, want 1 invocation (launchctl load)", runner.calls)
	}
	call := runner.calls[0]
	if call[0] != "launchctl" || call[1] != "load" {
		t.Errorf("call = %v, want launchctl load ...", call)
	}
}

func TestServiceInstallerInstallSkipsEnableWhenNotEnabled(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("service file not written even though enable=false: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Errorf("Run() calls = %v, want none when enable=false", runner.calls)
	}
}

func TestServiceInstallerInstallRejectsUnsupportedOS(t *testing.T) {
	t.Parallel()

	installer := ServiceInstaller{GOOS: "windows", HomeDir: t.TempDir(), Run: (&fakeRunner{}).run}
	_, err := installer.Install(testServiceConfig(), true)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Install() error = %v, want unsupported-platform error", err)
	}
}

func TestServiceInstallerUninstallRemovesFileAndStopsService(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	runner.calls = nil // reset; only care about uninstall's own calls

	found, gotPath, err := installer.Uninstall(true)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if !found {
		t.Error("Uninstall() found = false, want true")
	}
	if gotPath != path {
		t.Errorf("Uninstall() path = %q, want %q", gotPath, path)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("service file still exists after Uninstall(): %v", statErr)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("Run() calls = %v, want 1 invocation (disable --now)", runner.calls)
	}
	if !strings.Contains(strings.Join(runner.calls[0], " "), "disable --now okf-mcp.service") {
		t.Errorf("call = %v, want systemctl --user disable --now okf-mcp.service", runner.calls[0])
	}
}

func TestServiceInstallerUninstallToleratesStopFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	runner.queue("", errors.New("unit not loaded"))

	found, _, err := installer.Uninstall(true)
	if err != nil {
		t.Fatalf("Uninstall() error = %v, want nil (a failed stop shouldn't block removing the file)", err)
	}
	if !found {
		t.Error("Uninstall() found = false, want true")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("service file still exists after Uninstall(): %v", statErr)
	}
}

func TestServiceInstallerUninstallReportsNotFoundWhenNeverInstalled(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: t.TempDir(), Run: runner.run}

	found, _, err := installer.Uninstall(true)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if found {
		t.Error("Uninstall() found = true, want false")
	}
	if len(runner.calls) != 0 {
		t.Errorf("Run() calls = %v, want none when nothing was installed", runner.calls)
	}
}

func TestServiceInstallerStatusReportsNotInstalled(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: t.TempDir(), Run: runner.run}

	status, err := installer.Status(true)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Installed {
		t.Error("Status().Installed = true, want false")
	}
	if len(runner.calls) != 0 {
		t.Errorf("Run() calls = %v, want none when not installed", runner.calls)
	}
}

func TestServiceInstallerStatusRunsQueryWhenInstalled(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	path, err := installer.Install(testServiceConfig(), false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	runner.calls = nil
	runner.queue("active\n", nil)

	status, err := installer.Status(true)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Installed {
		t.Error("Status().Installed = false, want true")
	}
	if status.Path != path {
		t.Errorf("Status().Path = %q, want %q", status.Path, path)
	}
	if status.State != "active" {
		t.Errorf("Status().State = %q, want %q", status.State, "active")
	}
}

func TestServiceInstallerStatusDegradesGracefullyWhenQueryFails(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	runner := &fakeRunner{}
	installer := ServiceInstaller{GOOS: "linux", HomeDir: home, Run: runner.run}

	if _, err := installer.Install(testServiceConfig(), false); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	runner.calls = nil
	runner.queue("", errors.New("systemctl: command not found"))

	status, err := installer.Status(true)
	if err != nil {
		t.Fatalf("Status() error = %v, want nil (a failed live query shouldn't fail Status)", err)
	}
	if !status.Installed {
		t.Error("Status().Installed = false, want true (file exists regardless of query outcome)")
	}
	if status.State == "" {
		t.Error("Status().State is empty, want a note that the live query failed")
	}
}
