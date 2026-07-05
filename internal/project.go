package internal

import (
	"context"
	"fmt"
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
	applyForkOverrides(project, f.Parent.Name, f.WorkingDir)
	return project, nil
}

func forkName(parentName, workingDir string) string {
	return fmt.Sprintf("%s_%s", parentName, filepath.Base(workingDir))
}

func applyForkOverrides(project *types.Project, parentName, workingDir string) {
	for _, srv := range project.Services {
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

		srv.CustomLabels[COMPOSEFORK_LABEL] = workingDir

		project.Services[srv.Name] = srv
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
