package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (a *App) Cache() error {
	a.Log.Println("Creating a volume cache...")

	// We change directory to the project root so we always prepare caches based
	// off the main worktree. A child worktree could've made changes to the dependencies
	commonDir, err := projectRoot()
	if err != nil {
		return err
	}
	rootDir := strings.TrimSuffix(commonDir, ".git")
	err = os.Chdir(rootDir)
	if err != nil {
		return err
	}
	a.Log.Println("Switched dir to", rootDir)

	randomName := namesgenerator.GetRandomName(0)

	project, err := NewProject(randomName)
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	composeProject, err := project.Load()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	err = a.Compose.Build(context.Background(), composeProject, api.BuildOptions{})
	if err != nil {
		return fmt.Errorf("error building project: %w", err)
	}

	err = a.Compose.Create(context.Background(), composeProject, api.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	err = a.Compose.Up(context.Background(), composeProject, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{Wait: true}})

	err = a.Compose.Stop(context.Background(), composeProject.Name, api.StopOptions{})
	if err != nil {
		return fmt.Errorf("error stopping project: %w", err)
	}

	dir, err := cacheDir()
	if err != nil {
		return err
	}

	tarballPaths, err := a.exportVolumes(dir, composeProject)
	if err != nil {
		return err
	}

	err = a.Compose.Down(context.Background(), composeProject.Name, api.DownOptions{
		Volumes:       true,
		RemoveOrphans: true,
		Images:        "local",
	})
	if err != nil {
		return fmt.Errorf("error tearing down project: %w", err)
	}

	withDirLock(dir, func() error {
		for _, tarballPath := range tarballPaths {
			newPath := strings.ReplaceAll(tarballPath, randomName, project.Parent.Name)
			err := os.Rename(tarballPath, newPath)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func cacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, APP_NAME)
	return dir, nil
}

func (a *App) exportVolumes(dir string, project *types.Project) ([]string, error) {
	results := []string{}

	dir, err := cacheDir()
	if err != nil {
		return results, fmt.Errorf("error copying volumes: %w", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return results, fmt.Errorf("error copying volumes: %w", err)
	}

	binds := []string{fmt.Sprintf("%s:/out", dir)}
	commands := []string{"sh", "-c", `
    set -e
    echo "all args $@"
    while [ "$#" -ge 2 ]; do
      echo "copying $1 to $2"
      tar -c --numeric-owner -f $2 -C $1 .
      shift 2
    done
  `, "foo"}
	for _, vol := range project.Volumes {
		binds = append(binds, fmt.Sprintf("%s:/in/%s:ro", vol.Name, vol.Name))
		tarballName := fmt.Sprintf("%s.tar", vol.Name)
		inputPath := filepath.Join("/in", vol.Name)
		outputPath := filepath.Join("/out", tarballName)
		commands = append(commands, inputPath, outputPath)
		results = append(results, filepath.Join(dir, tarballName))
	}

	id, err := a.createSystemContainer(projectLabels(project.Name, project.WorkingDir), client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: commands,
		},
		HostConfig: &container.HostConfig{
			Binds:      binds,
			AutoRemove: true,
			UsernsMode: "host",
		},
	})
	if err != nil {
		return results, err
	}

	waitResult := a.Client().ContainerWait(context.Background(), id, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitResult.Error:
		return results, err
	case status := <-waitResult.Result:
		if status.StatusCode != 0 {
			return results, fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	a.Log.Println("Finished caching")
	return results, nil
}

func withDirLock(dir string, f func() error) error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := syscall.Flock(int(d.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(d.Fd()), syscall.LOCK_UN)

	return f()
}
