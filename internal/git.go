package internal

import (
	"errors"
	"os/exec"
	"strings"
)

// copyWorktreeIncludes seeds destRoot with the local files git left out of it:
// the entries .worktreeinclude names in sourceRoot, read as gitignore-style
// patterns and matched against the files git reports as untracked and ignored.
func copyWorktreeIncludes(sourceRoot, destRoot string) error {
	return errors.New("copyWorktreeIncludes: not implemented")
}

func inMainWorktree() (bool, error) {
	currentDir, err := currentRoot()
	if err != nil {
		return false, err
	}

	commonDir, err := projectRoot()
	if err != nil {
		return false, err
	}

	return currentDir == commonDir, nil
}

func currentRoot() (string, error) {
	gitDir, err := git("rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return gitDir, nil
}

func projectRoot() (string, error) {
	commonDir, err := git("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(commonDir, "/.git"), nil
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
