package integration

import (
	"os"
	"os/exec"
	"testing"
)

func TestRecordDemo(t *testing.T) {
	if os.Getenv("RECORD_DEMO") != "1" {
		t.Skip("Set RECORD_DEMO=1 to record the demo GIF")
	}

	root := projectRoot()

	// Build the demo Docker image
	build := exec.Command("docker", "build", "-f", "demo.Dockerfile", "-t", "temporal-lsp-demo", ".")
	build.Dir = root
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("docker build failed: %v", err)
	}

	// Run vhs inside the container and export the GIF
	run := exec.Command("docker", "run", "--rm", "--entrypoint", "sh",
		"-v", root+":/out",
		"temporal-lsp-demo",
		"-c", "vhs demo.tape && cp demo.gif /out/demo.gif",
	)
	run.Dir = root
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		t.Fatalf("docker run failed: %v", err)
	}

	// Verify the GIF was exported
	info, err := os.Stat(root + "/demo.gif")
	if err != nil {
		t.Fatalf("demo.gif not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("demo.gif is empty (0 bytes)")
	}

	t.Logf("demo.gif exported (%d bytes)", info.Size())
}
