package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests pin the semantics of .worktreeinclude: a worktree an agent made
// for itself with a plain `git worktree add` has to end up holding the same
// local files it would have been given otherwise.
//
// The load-bearing decision is that candidates come from git, not from walking
// the filesystem: `git ls-files --others --ignored --exclude-standard` lists
// only files git has already decided not to carry into a worktree. Tracked files
// are therefore never copy candidates and never collide, and the entries are
// gitignore-style patterns rather than literal paths.

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func writeUnder(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("creating %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %q: %v", path, err)
	}
}

// newIncludeFixture builds a main checkout and a linked worktree off it. tracked
// files are committed, so git checks them into the worktree; local files are
// written afterwards and never committed, so the worktree starts without them —
// which is exactly the gap .worktreeinclude exists to close. include of nil
// writes no .worktreeinclude at all.
func newIncludeFixture(t *testing.T, tracked, local map[string]string, gitignore, include []string) (mainDir, forkDir string) {
	t.Helper()

	mainDir = t.TempDir()
	mustGit(t, mainDir, "init", "-q", ".")

	for rel, content := range tracked {
		writeUnder(t, mainDir, rel, content)
	}
	if len(gitignore) > 0 {
		writeUnder(t, mainDir, ".gitignore", strings.Join(gitignore, "\n")+"\n")
	}
	if include != nil {
		writeUnder(t, mainDir, ".worktreeinclude", strings.Join(include, "\n")+"\n")
	}
	mustGit(t, mainDir, "add", "-A")
	mustGit(t, mainDir, "commit", "-qm", "project")

	// Written after the commit so they stay untracked, and gitignored so git
	// reports them as candidates.
	for rel, content := range local {
		writeUnder(t, mainDir, rel, content)
	}

	forkDir = filepath.Join(t.TempDir(), "fork")
	mustGit(t, mainDir, "worktree", "add", "-q", "--detach", forkDir)

	return mainDir, forkDir
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("reading %q: %v", path, err)
		return
	}
	if string(b) != want {
		t.Errorf("%q = %q, want %q", path, b, want)
	}
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Errorf("expected %q to be absent, it exists", path)
	}
}

func TestCopyWorktreeIncludes(t *testing.T) {
	cases := []struct {
		name       string
		tracked    map[string]string
		local      map[string]string
		gitignore  []string
		include    []string
		wantFork   map[string]string
		wantAbsent []string
	}{
		{
			// The case a filesystem walk cannot handle: git already checked
			// storage/.keep into the fork, so a whole-tree copy collides on it
			// before reaching the one file that was actually missing.
			name:      "tracked files are never candidates",
			tracked:   map[string]string{"storage/.keep": "", "config/application.rb": "app"},
			local:     map[string]string{"storage/local.sqlite3": "db"},
			gitignore: []string{"storage/local.sqlite3"},
			include:   []string{"storage/"},
			wantFork: map[string]string{
				"storage/local.sqlite3": "db",
				"storage/.keep":         "",
				"config/application.rb": "app",
			},
		},
		{
			// A wholly ignored directory comes back from ls-files collapsed to
			// "local/", so it has to be expanded before anything is copied. The
			// two directories are the two ways that can be got wrong: local/ is
			// named outright, while secrets/ is reached only by a pattern that
			// matches the files underneath it — which means the expansion has to
			// happen before the matching, not after. Neither directory exists in
			// the fork, so the parents are created on the way.
			name: "ignored directories are expanded and their parents created",
			local: map[string]string{
				"local/a.txt":        "a",
				"local/b.txt":        "b",
				"secrets/deep/a.key": "k",
			},
			gitignore: []string{"local/", "secrets/"},
			include:   []string{"local/", "**/*.key"},
			wantFork: map[string]string{
				"local/a.txt":        "a",
				"local/b.txt":        "b",
				"secrets/deep/a.key": "k",
			},
		},
		{
			// One .worktreeinclude carrying a comment, a blank line, a glob that
			// matches at any depth, and a negation. None of it survives being
			// read as a list of literal paths, which is what the first version of
			// this did. junk.txt is the trap: it is ignored, so it is a
			// candidate, and only the patterns keep it out — a blank line taken
			// to mean "everything" would pull it in.
			name: "entries are gitignore patterns, not literal paths",
			local: map[string]string{
				"config/local.key": "secret",
				"local/a.txt":      "a",
				"local/b.txt":      "b",
				"junk.txt":         "junk",
			},
			gitignore: []string{"*.key", "local/", "junk.txt"},
			// local/* rather than local/, because gitignore cannot re-include a
			// file whose parent directory the same list excluded.
			include: []string{"# local files", "", "   ", "*.key", "local/*", "!local/b.txt"},
			wantFork: map[string]string{
				"config/local.key": "secret",
				"local/a.txt":      "a",
			},
			wantAbsent: []string{"local/b.txt", "junk.txt"},
		},
		{
			// Untracked is not enough: git has to be ignoring it too, or the
			// file is something the developer simply has not committed yet.
			// Nothing here is ignored, so there are no candidates at all.
			name:       "untracked but unignored files are not copied",
			local:      map[string]string{"scratch.txt": "wip"},
			include:    []string{"scratch.txt"},
			wantAbsent: []string{"scratch.txt"},
		},
		{
			// The same rule with the candidate list non-empty, which is the case
			// that actually exercises it: .env keeps the list from being empty,
			// so scratch.txt is held back by the candidates restricting the
			// match rather than by there being nothing to match against.
			name:       "an unignored file is held back even when others are copied",
			local:      map[string]string{".env": "ENV", "scratch.txt": "wip"},
			gitignore:  []string{".env"},
			include:    []string{".env", "scratch.txt"},
			wantFork:   map[string]string{".env": "ENV"},
			wantAbsent: []string{"scratch.txt"},
		},
		{
			name:       "missing .worktreeinclude is a no-op",
			local:      map[string]string{".env": "ENV"},
			gitignore:  []string{".env"},
			include:    nil,
			wantAbsent: []string{".env"},
		},
		{
			// A separate branch from the missing file: this one is read, and
			// leaves no patterns behind. No patterns must not become every file.
			name:       "a file of only comments is a no-op",
			local:      map[string]string{".env": "ENV"},
			gitignore:  []string{".env"},
			include:    []string{"# nothing here", ""},
			wantAbsent: []string{".env"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mainDir, forkDir := newIncludeFixture(t, c.tracked, c.local, c.gitignore, c.include)

			if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
				t.Fatalf("copyWorktreeIncludes: %v", err)
			}

			for rel, want := range c.wantFork {
				assertContent(t, filepath.Join(forkDir, rel), want)
			}
			for _, rel := range c.wantAbsent {
				assertAbsent(t, filepath.Join(forkDir, rel))
			}
		})
	}
}

// A symlink's target is the developer's own machine, so copying it into a fork
// either duplicates something large or points at a path that means nothing
// there. Skipped, rather than failing the copy that surrounds it — which is what
// asserting real.txt landed is for.
func TestCopyWorktreeIncludesSkipsSymlinks(t *testing.T) {
	mainDir, forkDir := newIncludeFixture(t,
		nil,
		map[string]string{"local/real.txt": "real"},
		[]string{"local/"},
		[]string{"local/"},
	)
	if err := os.Symlink("real.txt", filepath.Join(mainDir, "local", "link.txt")); err != nil {
		t.Fatalf("creating symlink: %v", err)
	}

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("copyWorktreeIncludes: %v", err)
	}

	assertContent(t, filepath.Join(forkDir, "local/real.txt"), "real")
	assertAbsent(t, filepath.Join(forkDir, "local/link.txt"))
}

// Seeding is per file and once only. There is no creation hook to hang this off,
// so it runs on every up, and only ever adding what is absent is what makes a
// repeated call mean the same thing as a single one.
//
// The two edits are the two directions that can go wrong in. A developer who
// repointed the fork's .env would lose the edit every time the fork came up; and
// a fork restamped from a parent that has since moved on would be handed content
// it never asked for. Both are worse than doing nothing.
func TestCopyWorktreeIncludesNeverOverwrites(t *testing.T) {
	mainDir, forkDir := newIncludeFixture(t,
		nil,
		map[string]string{".env": "PARENT", "local/a.txt": "parent-a"},
		[]string{".env", "local/"},
		[]string{".env", "local/"},
	)

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("first copy: %v", err)
	}
	assertContent(t, filepath.Join(forkDir, ".env"), "PARENT")

	writeUnder(t, forkDir, ".env", "FORK-EDITED")
	writeUnder(t, mainDir, "local/a.txt", "parent-moved-on")

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("second copy: %v", err)
	}
	assertContent(t, filepath.Join(forkDir, ".env"), "FORK-EDITED")
	assertContent(t, filepath.Join(forkDir, "local/a.txt"), "parent-a")
}

// The other side of the same rule: absent means absent, whoever made it so. A
// fork that deletes its .env gets a fresh one on the next up rather than coming
// up without the project name, which is the failure this whole copy exists to
// prevent.
func TestCopyWorktreeIncludesReseedsDeletedFiles(t *testing.T) {
	mainDir, forkDir := newIncludeFixture(t,
		nil,
		map[string]string{".env": "PARENT"},
		[]string{".env"},
		[]string{".env"},
	)

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("first copy: %v", err)
	}
	if err := os.Remove(filepath.Join(forkDir, ".env")); err != nil {
		t.Fatalf("removing fork .env: %v", err)
	}

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("second copy: %v", err)
	}
	assertContent(t, filepath.Join(forkDir, ".env"), "PARENT")
}

// The ordinary case, and the one that pins the granularity: the app made the
// worktree and seeded it at creation, then the agent runs up in it. There is no
// way to detect that from inside the worktree, so the copy runs regardless and
// has to leave the seeded file alone while still filling the gap beside it. A
// seed can stop half way — an entry that was a symlink, a file that could not be
// read, a .worktreeinclude that gained a line after the worktree was made.
//
// The pre-seeded content is deliberately not the parent's: identical content
// would pass whether the file was skipped or restamped.
func TestCopyWorktreeIncludesCompletesAPartialSeed(t *testing.T) {
	mainDir, forkDir := newIncludeFixture(t,
		nil,
		map[string]string{".env": "PARENT", "local/a.txt": "parent-a"},
		[]string{".env", "local/"},
		[]string{".env", "local/"},
	)
	writeUnder(t, forkDir, ".env", "ALREADY-SEEDED")

	if err := copyWorktreeIncludes(mainDir, forkDir); err != nil {
		t.Fatalf("copyWorktreeIncludes: %v", err)
	}

	assertContent(t, filepath.Join(forkDir, ".env"), "ALREADY-SEEDED")
	assertContent(t, filepath.Join(forkDir, "local/a.txt"), "parent-a")
}

// A symlink in the fork points wherever it was made to point, so an include
// underneath one lands outside the worktree and writes over whatever is already
// there. The destination is resolved before anything is opened.
//
// z-local.txt sorts after the offending entry, so it also pins the other half of
// the rule: a bad entry is collected and the remaining ones are still seeded.
func TestCopyWorktreeIncludesRefusesToWriteOutsideTheWorktree(t *testing.T) {
	mainDir, forkDir := newIncludeFixture(t,
		nil,
		map[string]string{"uploads/secret.txt": "s", "z-local.txt": "z"},
		[]string{"uploads/", "z-local.txt"},
		[]string{"uploads/", "z-local.txt"},
	)

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(forkDir, "uploads")); err != nil {
		t.Fatalf("creating symlink: %v", err)
	}

	err := copyWorktreeIncludes(mainDir, forkDir)
	if err == nil {
		t.Fatal("expected an error naming the entry that escaped, got nil")
	}
	if !strings.Contains(err.Error(), "uploads/secret.txt") {
		t.Errorf("error %q does not name the offending entry", err)
	}

	assertAbsent(t, filepath.Join(outside, "secret.txt"))
	assertContent(t, filepath.Join(forkDir, "z-local.txt"), "z")
}
