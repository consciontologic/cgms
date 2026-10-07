package sim

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPrivateWrapperRedactsSuccessAndFailure(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	base := filepath.Join(root, "sims/artifacts/privacy-tests")
	if e = os.MkdirAll(base, 0700); e != nil {
		t.Fatal(e)
	}
	dir, e := os.MkdirTemp(base, "case-")
	if e != nil {
		t.Fatal(e)
	}
	defer os.RemoveAll(dir)
	for _, exit := range []string{"0", "5"} {
		spec := map[string]any{"argv": []string{"python3", "-c", "import sys; print('PRIVATE_CANARY_SEED_7632'); sys.exit(" + exit + ")"}, "cwd": root, "capture": filepath.Join(dir, "capture-"+exit)}
		b, _ := json.Marshal(spec)
		path := filepath.Join(dir, "request-"+exit+".json")
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
		cmd := exec.Command("python3", filepath.Join(root, "xops/sim_private.py"), path)
		output, err := cmd.CombinedOutput()
		if strings.Contains(string(output), "PRIVATE_CANARY") {
			t.Fatal("private child output leaked")
		}
		if exit == "0" && err != nil {
			t.Fatalf("helper unavailable: %s", output)
		}
		if exit == "5" && (err == nil || cmd.ProcessState.ExitCode() != 5) {
			t.Fatal("lost exit status")
		}
		private, e := os.ReadFile(spec["capture"].(string) + ".log")
		if e != nil || !strings.Contains(string(private), "PRIVATE_CANARY") {
			t.Fatal("private forensic output missing")
		}
	}
}

func privateTestRoot(t *testing.T) string {
	t.Helper()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	base := filepath.Join(root, "sims/artifacts/privacy-tests")
	if e = os.MkdirAll(base, 0700); e != nil {
		t.Fatal(e)
	}
	dir, e := os.MkdirTemp(base, "isolated-")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}
func TestPrivateWrapperRejectsLexicalEscape(t *testing.T) {
	root, _ := filepath.Abs("../../..")
	dir := privateTestRoot(t)
	probe := "privacy-escape-probe-" + filepath.Base(dir)
	capture := filepath.Join(root, "sims/artifacts") + "/../" + probe
	spec := map[string]any{"argv": []string{"python3", "-c", "print('PRIVATE_ESCAPE_CANARY')"}, "cwd": root, "capture": capture}
	raw, _ := json.Marshal(spec)
	path := filepath.Join(dir, "spec.json")
	os.WriteFile(path, raw, 0600)
	cmd := exec.Command("python3", filepath.Join(root, "xops/sim_private.py"), path)
	output, e := cmd.CombinedOutput()
	for _, suffix := range []string{".log", ".argv.json"} {
		escaped := filepath.Join(root, "sims", probe) + suffix
		if _, err := os.Stat(escaped); err == nil {
			os.Remove(escaped)
			t.Errorf("protected capture escaped artifact root")
		}
	}
	if e == nil {
		t.Fatal("lexical escape accepted")
	}
	if strings.Contains(string(output), "PRIVATE_ESCAPE_CANARY") {
		t.Fatal("private diagnostic leaked")
	}
}
func TestSafeRunPrivateWrapperRedactsOuterForensics(t *testing.T) {
	root, _ := filepath.Abs("../../..")
	dir := privateTestRoot(t)
	fake := filepath.Join(dir, "repo")
	for _, relative := range []string{"xops/agent/safe-run.sh", "xops/lib/log.sh", "xops/sim_private.py"} {
		src, e := os.ReadFile(filepath.Join(root, relative))
		if e != nil {
			t.Fatal(e)
		}
		dest := filepath.Join(fake, relative)
		os.MkdirAll(filepath.Dir(dest), 0700)
		if e = os.WriteFile(dest, src, 0700); e != nil {
			t.Fatal(e)
		}
	}
	artifactRoot := filepath.Join(fake, "sims/artifacts")
	os.MkdirAll(artifactRoot, 0700)
	for _, status := range []int{0, 5} {
		capture := filepath.Join(artifactRoot, fmt.Sprintf("private-%d", status))
		spec := map[string]any{"argv": []string{"python3", "-c", fmt.Sprintf("import sys; print('PRIVATE_OUTER_CANARY_8821'); sys.exit(%d)", status)}, "cwd": fake, "capture": capture}
		b, _ := json.Marshal(spec)
		specPath := filepath.Join(artifactRoot, fmt.Sprintf("request-%d.json", status))
		os.WriteFile(specPath, b, 0600)
		logs := filepath.Join(fake, fmt.Sprintf("safe-logs-%d", status))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "bash", filepath.Join(fake, "xops/agent/safe-run.sh"), "privacy-boundary", "--", "python3", filepath.Join(fake, "xops/sim_private.py"), specPath)
		cmd.Dir = fake
		cmd.Env = append(os.Environ(), "AVB_RUN_DIR="+logs, "AVB_HEARTBEAT_SECS=1")
		output, e := cmd.CombinedOutput()
		cancel()
		if status == 0 && e != nil || status != 0 && (e == nil || cmd.ProcessState.ExitCode() != status) {
			t.Fatalf("outer exit lost: %v %s", e, output)
		}
		if strings.Contains(string(output), "PRIVATE_OUTER_CANARY") {
			t.Fatal("safe-run terminal leak")
		}
		entries, e := os.ReadDir(logs)
		if e != nil {
			t.Fatal(e)
		}
		if len(entries) != 3 {
			t.Fatal("missing safe-run forensics")
		}
		for _, entry := range entries {
			data, e := os.ReadFile(filepath.Join(logs, entry.Name()))
			if e != nil {
				t.Fatal(e)
			}
			if strings.Contains(string(data), "PRIVATE_OUTER_CANARY") {
				t.Fatal("outer forensic leak", entry.Name())
			}
		}
		if status != 0 {
			failure, e := os.ReadFile(filepath.Join(fake, "docs/tracking/state/last_failure.json"))
			if e != nil {
				t.Fatal(e)
			}
			if strings.Contains(string(failure), "PRIVATE_OUTER_CANARY") {
				t.Fatal("failure breadcrumb leak")
			}
		}
		private, e := os.ReadFile(capture + ".log")
		if e != nil || !strings.Contains(string(private), "PRIVATE_OUTER_CANARY") {
			t.Fatal("private forensic evidence missing")
		}
	}
}
func TestPrivateWrapperInterruptBounded(t *testing.T) {
	root, _ := filepath.Abs("../../..")
	dir := privateTestRoot(t)
	ready := filepath.Join(dir, "ready")
	spec := map[string]any{"argv": []string{"python3", "-c", "import signal,sys,time,pathlib; signal.signal(signal.SIGINT, lambda *args: sys.exit(130)); pathlib.Path(sys.argv[1]).write_text('ready'); time.sleep(30)", ready}, "cwd": root, "capture": filepath.Join(dir, "interrupt")}
	b, _ := json.Marshal(spec)
	path := filepath.Join(dir, "request.json")
	os.WriteFile(path, b, 0600)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", filepath.Join(root, "xops/sim_private.py"), path)
	if e := cmd.Start(); e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, e := os.Stat(ready); e == nil {
			break
		}
		if time.Now().After(deadline) {
			cmd.Process.Signal(syscall.SIGINT)
			cmd.Wait()
			t.Fatal("child did not become ready")
		}
		time.Sleep(5 * time.Millisecond)
	}
	started := time.Now()
	if e := cmd.Process.Signal(syscall.SIGINT); e != nil {
		t.Fatal(e)
	}
	e := cmd.Wait()
	if e == nil || cmd.ProcessState.ExitCode() != 130 || time.Since(started) > time.Second {
		t.Fatal("graceful simulator cancellation not bounded", e)
	}
}

func TestPrivateWrapperStrictSpecification(t *testing.T) {
	root, _ := filepath.Abs("../../..")
	dir := privateTestRoot(t)
	valid := map[string]any{"argv": []string{"python3", "-c", "print('PRIVATE_STRICT_CANARY')"}, "cwd": root, "capture": filepath.Join(dir, "private")}
	raw, _ := json.Marshal(valid)
	cases := map[string][]byte{"duplicate": []byte(`{"argv":[],"argv":["PRIVATE_STRICT_CANARY"],"cwd":"x","capture":"x"}`), "unknown": []byte(`{"argv":[],"cwd":"x","capture":"x","secret":"PRIVATE_STRICT_CANARY"}`), "wrong-type": []byte(`{"argv":["python3"],"cwd":[],"capture":"x"}`), "malformed": []byte(`{"PRIVATE_STRICT_CANARY"`)}
	for name, b := range cases {
		path := filepath.Join(dir, name+".json")
		os.WriteFile(path, b, 0600)
		cmd := exec.Command("python3", filepath.Join(root, "xops/sim_private.py"), path)
		out, e := cmd.CombinedOutput()
		if e == nil || cmd.ProcessState.ExitCode() != 2 || strings.Contains(string(out), "PRIVATE_STRICT_CANARY") {
			t.Fatal("invalid specification accepted or disclosed", name, e, string(out))
		}
	}
	target := filepath.Join(dir, "valid.json")
	os.WriteFile(target, raw, 0600)
	link := filepath.Join(dir, "link.json")
	if e := os.Symlink(target, link); e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command("python3", filepath.Join(root, "xops/sim_private.py"), link)
	if out, e := cmd.CombinedOutput(); e == nil || cmd.ProcessState.ExitCode() != 2 || strings.Contains(string(out), "PRIVATE_STRICT_CANARY") {
		t.Fatal("symlink specification accepted or disclosed")
	}
	os.Chmod(target, 0644)
	cmd = exec.Command("python3", filepath.Join(root, "xops/sim_private.py"), target)
	if e := cmd.Run(); e == nil || cmd.ProcessState.ExitCode() != 2 {
		t.Fatal("publicly readable private specification accepted")
	}
}
