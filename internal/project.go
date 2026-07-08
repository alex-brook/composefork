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

type Fork struct {
	Parent     *types.Project // the compose project as authored
	Name       string         // fork project name: {Parent.Name}_{basename(WorkingDir)}
	WorkingDir string         // absolute working directory, resolved once
}

func LoadFork() (*Fork, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	parent, err := loadComposeProject("")
	if err != nil {
		return nil, err
	}

	return &Fork{
		Parent:     parent,
		Name:       forkName(parent.Name, wd),
		WorkingDir: wd,
	}, nil
}

func (f *Fork) Project() (*types.Project, error) {
	project, err := loadComposeProject(f.Name)
	if err != nil {
		return nil, err
	}
	applyForkOverrides(project, f.Parent.Name, f.Labels())
	return project, nil
}

func projectLabels(projectName, dir string) map[string]string {
	return map[string]string{
		COMPOSEFORK_PROJECT_LABEL: projectName, // shared group id / discovery marker
		COMPOSEFORK_DIR_LABEL:     dir,         // dir path (prune's existence check)
	}
}

func (f *Fork) Labels() map[string]string {
	return projectLabels(f.Parent.Name, f.WorkingDir)
}

func forkName(parentName, workingDir string) string {
	return fmt.Sprintf("%s_%s", parentName, filepath.Base(workingDir))
}

func applyForkOverrides(project *types.Project, parentName string, labels map[string]string) {
	for name, srv := range project.Services {
		if srv.Image == "" {
			srv.Image = fmt.Sprintf("%s-%s:%s", parentName, srv.Name, "latest")
			srv.Build = &types.BuildConfig{}
		}
		for i := range srv.Ports {
			srv.Ports[i].Published = ""
			srv.Ports[i].HostIP = ""
		}
		if srv.CustomLabels == nil {
			srv.CustomLabels = types.Labels{}
		}
		// Set expected compose labels for this container
		srv.CustomLabels[api.ProjectLabel] = project.Name
		srv.CustomLabels[api.ServiceLabel] = srv.Name
		srv.CustomLabels[api.OneoffLabel] = "False"
		srv.CustomLabels[api.WorkingDirLabel] = project.WorkingDir
		srv.CustomLabels[api.ConfigFilesLabel] = strings.Join(project.ComposeFiles, ",")

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
