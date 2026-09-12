package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyWorktreeIncludes(sourceRoot, destRoot string) error {
	patternFile := filepath.Join(sourceRoot, ".worktreeinclude")
	if _, err := os.Stat(patternFile); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	// Only files git already left out of a worktree are candidates, so a tracked
	// file can never collide. A wholly ignored directory arrives as "dir/"
	candidates, err := lsFiles(sourceRoot, "--exclude-standard", "--directory")
	if err != nil {
		return err
	}

	// An empty pathspec below would mean "no restriction", not "nothing"
	if len(candidates) == 0 {
		return nil
	}

	// git applies the patterns and expands the collapsed directories in one pass.
	// :(literal) stops a candidate named "*.env" being read as a glob
	args := []string{"--exclude-from=" + patternFile, "--"}
	for _, candidate := range candidates {
		args = append(args, ":(literal)"+candidate)
	}
	includes, err := lsFiles(sourceRoot, args...)
	if err != nil {
		return err
	}

	// A committed symlink in the fork could redirect a write out of the worktree.
	// os.Root refuses to follow one, so containment is the kernel's business
	dest, err := os.OpenRoot(destRoot)
	if err != nil {
		return err
	}
	defer dest.Close()

	// One entry we can't copy shouldn't cost the fork the rest of them
	var errs []error
	for _, include := range includes {
		if err := seedFile(dest, filepath.Join(sourceRoot, include), include); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", include, err))
		}
	}
	return errors.Join(errs...)
}

func seedFile(dest *os.Root, sourcePath, rel string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}

	// A symlink's target is a path on the machine that made it
	if !info.Mode().IsRegular() {
		return nil
	}

	if err := dest.MkdirAll(filepath.Dir(rel), 0755); err != nil {
		return err
	}

	reader, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// O_EXCL rather than a stat first, so a file the fork already has is left
	// alone with no window between the check and the create
	writer, err := dest.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if errors.Is(err, fs.ErrExist) {
		return nil
	} else if err != nil {
		return err
	}

	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return err
	}

	// A write can still fail on close, reporting a truncated copy as a success
	return writer.Close()
}

func lsFiles(dir string, args ...string) ([]string, error) {
	out, err := git(append([]string{"-C", dir, "ls-files", "--others", "--ignored", "-z"}, args...)...)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, path := range strings.Split(out, "\x00") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths, nil
}

func inMainWorktree() (bool, error) {
	sourceRoot, destRoot, err := worktreeRoots()
	if err != nil {
		return false, err
	}

	return sourceRoot == destRoot, nil
}

func worktreeRoots() (sourceRoot, destRoot string, err error) {
	sourceRoot, err = projectRoot()
	if err != nil {
		return "", "", err
	}

	destRoot, err = currentRoot()
	if err != nil {
		return "", "", err
	}

	return sourceRoot, destRoot, nil
}

func currentRoot() (string, error) {
	return git("rev-parse", "--path-format=absolute", "--show-toplevel")
}

func projectRoot() (string, error) {
	commonDir, err := git("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(commonDir, "/.git"), nil
}

func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	// Output discards stderr, leaving a failure reading "exit status 128"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, message)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}
