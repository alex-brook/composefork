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
		{"app", "/home/user/wt", "app_wt"},
		{"app", "/home/user/wt/", "app_wt"}, // Base strips the trailing slash
		{"myproj", "/tmp/feature-x", "myproj_feature-x"},
	}
	for _, c := range cases {
		if got := forkName(c.parent, c.dir); got != c.want {
			t.Errorf("forkName(%q, %q) = %q, want %q", c.parent, c.dir, got, c.want)
		}
	}
}

func TestForkLabels(t *testing.T) {
	f := &Fork{Parent: &types.Project{Name: "app"}, WorkingDir: "/tmp/wt"}
	got := f.Labels()
	want := map[string]string{
		COMPOSEFORK_PROJECT_LABEL: "app",
		COMPOSEFORK_DIR_LABEL:     "/tmp/wt",
	}
	if !maps.Equal(got, want) {
		t.Errorf("Labels() = %v, want %v", got, want)
	}
}
