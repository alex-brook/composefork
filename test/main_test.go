package test

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	// The dummy's web service joins the external gem_net network, which the
	// gemstash project creates; gemstash is also a pull-through gem cache. Bring
	// it up for the whole suite, and stop it after (keeping the cache volume).
	if err := gemstash("up", "-d", "--wait"); err != nil {
		log.Printf("gemstash setup failed: %v", err)
		return 1 // all docker tests need gem_net; don't run without it
	}
	defer func() {
		if err := gemstash("down"); err != nil {
			log.Printf("gemstash teardown failed: %v", err)
		}
	}()

	return m.Run()
}

// gemstash runs `docker compose` against the gemstash project.
func gemstash(args ...string) error {
	full := append([]string{"compose", "-f", "gemstash/compose.yml"}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
