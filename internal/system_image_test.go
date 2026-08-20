package internal

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/moby/go-archive/compression"
)

// The bundled system image is a FROM scratch image holding one static binary,
// linked as /import and /export, which dispatches on argv[0]. These tests cover
// the image artifact on its own terms — that the embedded tarball is tagged, and
// that a filesystem survives an export/import round trip with its ownership and
// modes intact. The mounts and commands below mirror what exportVolumes and
// importVolumes construct, so they pin the image's contract rather than sharing
// code with the call sites.

// TestSystemImageHasTag guards the buildx `-t` flag. Without it the tarball
// loads untagged and createSystemContainer can't resolve SYSTEM_IMAGE — a
// failure that would otherwise only show up at runtime.
func TestSystemImageHasTag(t *testing.T) {
	tr := tar.NewReader(bytes.NewReader(systemImageTarball))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			t.Fatal("no manifest.json in embedded image tarball")
		}
		if err != nil {
			t.Fatalf("reading embedded image tarball: %v", err)
		}
		if hdr.Name != "manifest.json" {
			continue
		}

		var manifest []struct{ RepoTags []string }
		if err := json.NewDecoder(tr).Decode(&manifest); err != nil {
			t.Fatalf("decoding manifest.json: %v", err)
		}
		for _, m := range manifest {
			if slices.Contains(m.RepoTags, SYSTEM_IMAGE+":latest") {
				return
			}
		}
		t.Fatalf("embedded image tags = %v, want %s:latest", manifest, SYSTEM_IMAGE)
	}
}

func TestSystemImageRoundTrip(t *testing.T) {
	if err := exec.Command("docker", "version").Run(); err != nil {
		t.Skipf("docker unavailable: %v", err)
	}
	loadSystemImage(t)

	tmp := t.TempDir()
	// Created by the test user so cleanup can unlink the root-owned files the
	// container writes into it.
	data := filepath.Join(tmp, "data")
	if err := os.Mkdir(data, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureTar(t, filepath.Join(tmp, "src.tar"))

	runSystemImage(t,
		[]string{tmp + ":/cache", data + ":/out/data"},
		"import", "/cache/src.tar:/out/data")

	runSystemImage(t,
		[]string{data + ":/in/data:ro", tmp + ":/out"},
		"export", "/in/data:/out/dst.tar")

	got, want := readTar(t, filepath.Join(tmp, "dst.tar")), fixture()
	if !maps.Equal(got, want) {
		t.Errorf("round trip changed the archive\n got %+v\nwant %+v", got, want)
	}
}

// entry is the tar metadata the round trip must preserve. mtime is excluded
// deliberately — it survives fine but makes failures noisy.
type entry struct {
	typeflag byte
	uid, gid int
	mode     int64
	linkname string
	content  string
}

// fixture covers what the GNU `tar --numeric-owner -xp` invocation used to
// guarantee: non-root ownership, an owner that differs between a directory and
// its contents, a setuid bit, a symlink, and an empty directory.
func fixture() map[string]entry {
	return map[string]entry{
		"owned.txt": {typeflag: tar.TypeReg, uid: 1000, gid: 1000, mode: 0o644, content: "hello"},
		// World writable so t.TempDir cleanup, running as the test user, can
		// unlink the root-owned file this directory holds.
		"nested":        {typeflag: tar.TypeDir, uid: 999, gid: 999, mode: 0o777},
		"nested/setuid": {typeflag: tar.TypeReg, uid: 0, gid: 0, mode: 0o4755, content: "suid"},
		"empty":         {typeflag: tar.TypeDir, uid: 1000, gid: 1000, mode: 0o700},
		"link":          {typeflag: tar.TypeSymlink, uid: 1000, gid: 1000, linkname: "owned.txt"},
	}
}

func writeFixtureTar(t *testing.T, path string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	// Sorted, which puts a parent directory ahead of its contents since the
	// parent's path is a prefix of theirs.
	entries := fixture()
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		e := entries[name]
		hdr := &tar.Header{
			Name:     name,
			Typeflag: e.typeflag,
			Mode:     e.mode,
			Uid:      e.uid,
			Gid:      e.gid,
			Linkname: e.linkname,
			Size:     int64(len(e.content)),
			Format:   tar.FormatPAX,
		}
		if e.typeflag == tar.TypeDir {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
}

func readTar(t *testing.T, path string) map[string]entry {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("export produced no archive at %s: %v", path, err)
	}
	defer f.Close()

	// Export gzips despite the .tar name.
	r, err := compression.DecompressStream(f)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	got := map[string]entry{}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return got
		}
		if err != nil {
			t.Fatal(err)
		}

		name := filepath.Clean(hdr.Name)
		if name == "." {
			continue
		}

		e := entry{
			typeflag: hdr.Typeflag,
			uid:      hdr.Uid,
			gid:      hdr.Gid,
			mode:     hdr.Mode,
			linkname: hdr.Linkname,
		}
		if hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			e.content = string(content)
		}
		// Linux has no lchmod, so a symlink's mode is whatever the extracting
		// filesystem reports; only the link target is meaningful.
		if hdr.Typeflag == tar.TypeSymlink {
			e.mode = 0
		}
		got[name] = e
	}
}

// loadSystemImage loads the embedded tarball over stdin, so the test never
// needs to spill it to disk.
func loadSystemImage(t *testing.T) {
	t.Helper()

	cmd := exec.Command("docker", "load")
	cmd.Stdin = bytes.NewReader(systemImageTarball)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("docker load: %v\n%s", err, out)
	}
}

func runSystemImage(t *testing.T, mounts []string, cmd ...string) {
	t.Helper()

	args := []string{"run", "--rm", "--userns=host"} // userns as cache/up do, to pass uids through
	for _, m := range mounts {
		args = append(args, "-v", m)
	}
	args = append(args, SYSTEM_IMAGE)
	args = append(args, cmd...)

	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		t.Fatalf("docker %v: %v\n%s", args, err, out)
	}
}
