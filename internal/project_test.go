package internal

import (
	"maps"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestForkName(t *testing.T) {
	cases := []struct {
		parent, dir, want string
	}{
		{"app", "/home/user/wt", "app-wt"},
		{"app", "/home/user/wt/", "app-wt"}, // Base strips the trailing slash
		{"myproj", "/tmp/feature-x", "myproj-feature-x"},
	}
	for _, c := range cases {
		if got := forkName(c.parent, c.dir); got != c.want {
			t.Errorf("forkName(%q, %q) = %q, want %q", c.parent, c.dir, got, c.want)
		}
	}
}

func TestForkLabels(t *testing.T) {
	f := &Project{Parent: &types.Project{Name: "app"}, WorkingDir: "/tmp/wt"}
	got := f.Labels()
	want := map[string]string{
		COMPOSEFORK_PROJECT_LABEL: "app",
		COMPOSEFORK_DIR_LABEL:     "/tmp/wt",
	}
	if !maps.Equal(got, want) {
		t.Errorf("Labels() = %v, want %v", got, want)
	}
}

// Forks are agent scratch space and must not be reachable from the local
// network, so applyForkOverrides has to bind every published port to loopback —
// clearing HostIP is not enough, since a bare `3000:3000` leaves it empty and
// the daemon then binds 0.0.0.0.
func TestApplyForkOverridesBindsLoopback(t *testing.T) {
	cases := []struct {
		name  string
		ports []types.ServicePortConfig
	}{
		{"bare publish", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "tcp"},
		}},
		{"explicit loopback", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "tcp", HostIP: "127.0.0.1"},
		}},
		{"explicit wildcard", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "tcp", HostIP: "0.0.0.0"},
		}},
		{"lan address", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "tcp", HostIP: "192.168.1.5"},
		}},
		{"udp", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "udp"},
		}},
		// The loader expands ranges before we see them, so a range reaches
		// applyForkOverrides as one entry per port.
		{"expanded range", []types.ServicePortConfig{
			{Target: 3000, Published: "3000", Protocol: "tcp"},
			{Target: 3001, Published: "3001", Protocol: "tcp"},
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			project := &types.Project{Services: types.Services{
				"web": {Name: "web", Ports: c.ports},
			}}

			applyForkOverrides(project, "app")

			got := project.Services["web"].Ports
			if len(got) != len(c.ports) {
				t.Fatalf("port count = %d, want %d", len(got), len(c.ports))
			}
			for i, p := range got {
				if p.HostIP != "127.0.0.1" {
					t.Errorf("port %d (%d/%s): HostIP = %q, want %q", i, p.Target, p.Protocol, p.HostIP, "127.0.0.1")
				}
				// Published stays empty so the daemon assigns a free port and
				// parallel forks can't collide.
				if p.Published != "" {
					t.Errorf("port %d (%d/%s): Published = %q, want empty", i, p.Target, p.Protocol, p.Published)
				}
			}
		})
	}
}

// A service that publishes nothing must survive the override untouched.
func TestApplyForkOverridesNoPorts(t *testing.T) {
	project := &types.Project{Services: types.Services{
		"worker": {Name: "worker"},
	}}

	applyForkOverrides(project, "app")

	if got := project.Services["worker"].Ports; len(got) != 0 {
		t.Errorf("Ports = %v, want none", got)
	}
}
