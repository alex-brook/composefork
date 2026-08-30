package internal

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
)

type Project struct {
	Parent     *types.Project // the compose project as authored
	Name       string         // project name: parent's in the main worktree, {Parent.Name}-{basename(WorkingDir)} in a fork
	WorkingDir string         // absolute working directory, resolved once
}

func NewProject(customName string) (*Project, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	parent, err := loadComposeProject("")
	if err != nil {
		return nil, err
	}

	mainWorktree, err := inMainWorktree()
	if err != nil {
		return nil, err
	}

	var name string
	if customName != "" {
		name = customName
	} else if mainWorktree {
		name = parent.Name
	} else {
		name = forkName(parent.Name, wd)
	}

	return &Project{
		Parent:     parent,
		Name:       name,
		WorkingDir: wd,
	}, nil
}

func (p *Project) Root() bool {
	return p.Name == p.Parent.Name
}

func (p *Project) Load() (*types.Project, error) {
	project, err := loadComposeProject(p.Name)
	if err != nil {
		return nil, err
	}
	applyLabels(project, p.Labels())
	if !p.Root() {
		applyForkOverrides(project, p.Parent.Name)
	}

	return project, nil
}

func projectLabels(projectName, dir string) map[string]string {
	return map[string]string{
		COMPOSEFORK_PROJECT_LABEL: projectName, // shared group id / discovery marker
		COMPOSEFORK_DIR_LABEL:     dir,         // dir path (prune's existence check)
	}
}

func (p *Project) Labels() map[string]string {
	return projectLabels(p.Parent.Name, p.WorkingDir)
}

func forkName(parentName, workingDir string) string {
	return fmt.Sprintf("%s-%s", parentName, filepath.Base(workingDir))
}

func applyLabels(project *types.Project, labels map[string]string) {
	for name, srv := range project.Services {
		if srv.CustomLabels == nil {
			srv.CustomLabels = types.Labels{}
		}
		// Default labels
		srv.CustomLabels[api.ProjectLabel] = project.Name
		srv.CustomLabels[api.ServiceLabel] = srv.Name
		srv.CustomLabels[api.OneoffLabel] = "False"
		srv.CustomLabels[api.WorkingDirLabel] = project.WorkingDir
		srv.CustomLabels[api.ConfigFilesLabel] = strings.Join(project.ComposeFiles, ",")

		// Custom labels
		maps.Copy(srv.CustomLabels, labels)

		project.Services[name] = srv
	}
	for name, vol := range project.Volumes {
		if vol.CustomLabels == nil {
			vol.CustomLabels = types.Labels{}
		}
		maps.Copy(vol.CustomLabels, labels)
		project.Volumes[name] = vol
	}

	for name, net := range project.Networks {
		if net.CustomLabels == nil {
			net.CustomLabels = types.Labels{}
		}
		maps.Copy(net.CustomLabels, labels)
		project.Networks[name] = net
	}
}

func applyForkOverrides(project *types.Project, parentName string) {
	for name, srv := range project.Services {
		for i := range srv.Ports {
			srv.Ports[i].Published = ""
			srv.Ports[i].HostIP = "127.0.0.1"
		}
		project.Services[name] = srv
	}
}

func loadComposeProject(name string) (*types.Project, error) {
	opts, err := cli.NewProjectOptions(
		nil,
		cli.WithOsEnv,
		cli.WithEnvFiles(),
		cli.WithDotEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath,
		cli.WithResolvedPaths(true),
	)
	if err != nil {
		return nil, err
	}

	if name != "" {
		opts.Environment["COMPOSE_PROJECT_NAME"] = name
	}

	return opts.LoadProject(context.Background())
}
