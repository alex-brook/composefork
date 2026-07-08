package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

const APP_NAME = "composefork"
const COMPOSEFORK_PROJECT_LABEL = "com.github.alex-brook.composefork.project"
const COMPOSEFORK_DIR_LABEL = "com.github.alex-brook.composefork.dir"
const SYSTEM_IMAGE = "composefork/system"

func (a *App) printProjectStatus(name string) error {
	containers, err := a.Compose.Ps(context.Background(), name, api.PsOptions{All: false})
	if err != nil {
		return fmt.Errorf("error in ps: %w", err)
	}

	fmt.Printf("Project: %s\n", name)
	fmt.Printf("%-10s %-12s %-12s %s\n", "Service", "State", "Health", "Ports")
	for _, container := range containers {
		var ports strings.Builder
		published := 0
		for _, publisher := range container.Publishers {
			if publisher.PublishedPort == 0 {
				continue
			}
			if published >= 1 {
				ports.WriteString(", ")
			}
			portsLine := fmt.Sprintf("%s:%d->%d/%s", publisher.URL, publisher.PublishedPort, publisher.TargetPort, publisher.Protocol)
			ports.WriteString(portsLine)
			published += 1
		}

		fmt.Printf("%-10s %-12s %-12s %s\n", container.Service, container.State, container.Health, ports.String())
	}

	return nil
}

func (a *App) createSystemContainer(labels map[string]string, opts client.ContainerCreateOptions) (string, error) {
	docker := a.Client()

	// Load the bundled alpine image
	resp, err := docker.ImageLoad(context.Background(), bytes.NewReader(systemImageTarball))
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}
	defer resp.Close()
	io.Copy(io.Discard, resp)

	if opts.Config == nil {
		opts.Config = &container.Config{}
	}
	if opts.HostConfig == nil {
		opts.HostConfig = &container.HostConfig{}
	}

	opts.Config.Image = SYSTEM_IMAGE
	opts.Config.Labels = labels
	// opts.HostConfig.AutoRemove = true
	createResp, err := docker.ContainerCreate(context.Background(), opts)
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}
	_, err = docker.ContainerStart(context.Background(), createResp.ID, client.ContainerStartOptions{})

	return createResp.ID, nil
}
