package internal

import (
	"os/exec"
	"strings"
)

func inMainWorktree() (bool, error) {
	gitDir, err := git("rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return false, err
	}

	commonDir, err := projectRoot()
	if err != nil {
		return false, err
	}

	return gitDir == commonDir, nil
}

func projectRoot() (string, error) {
	commonDir, err := git("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return commonDir, nil
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
