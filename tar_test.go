package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"testing"
)

var samplePathSpecs = []struct {
	path       string
	isDir      bool
	linkTarget string
	mode       fs.FileMode
}{
	{"dir1", true, "", 0700},
	{path.Join("dir1", "a"), false, "", 0700},
	{path.Join("dir1", "b"), false, "", 0750},
	{"dir2", true, "", 0750},
	{path.Join("dir2", "c"), false, "", 0770},
	{path.Join("dir2", "d"), false, "", 0707},
	{path.Join("dir2", "e"), false, path.Join("some", "path"), 0},
}

// Creates the hierarchy of files/dirs described by samplePathSpecs
// and returns the top-level dirs. Calls Fatal on error.
func makeSamplePaths(sampleDir string) []string {
	os.RemoveAll(sampleDir)
	topDirs := []string{}
	for _, spec := range samplePathSpecs {
		fullPath := path.Join(sampleDir, spec.path)
		if spec.isDir {
			if err := os.MkdirAll(fullPath, spec.mode); err != nil {
				log.Fatal(err)
			}
			if spec.path == path.Base(spec.path) {
				topDirs = append(topDirs, spec.path)
			}
		} else if len(spec.linkTarget) > 0 {
			if err := os.Symlink(spec.linkTarget, fullPath); err != nil {
				log.Fatal(err)
			}
		} else {
			f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, spec.mode)
			if err != nil {
				log.Fatal(err)
			}
			_, err = f.Write([]byte(fmt.Sprintf("content of %v", spec.path)))
			if err != nil {
				log.Fatal(err)
			}
			f.Close()
		}
	}
	return topDirs
}

func validateSamplePaths(t *testing.T, dir string) {
	for _, spec := range samplePathSpecs {
		fullPath := path.Join(dir, spec.path)
		if spec.isDir {
			stat, err := os.Stat(fullPath)
			if err != nil {
				t.Errorf("error stating %v", fullPath)
				return
			}
			if !stat.IsDir() {
				t.Errorf("not a dir as expected: %v", fullPath)
			}
			if stat.Mode()&0777 != spec.mode {
				t.Errorf("unexpected mode for %v: %v != %v",
					fullPath, stat.Mode(), spec.mode)
			}
		} else if len(spec.linkTarget) > 0 {
			linkTarget, err := os.Readlink(fullPath)
			if err != nil {
				log.Fatal(err)
			}
			if linkTarget != spec.linkTarget {
				t.Errorf("unexpected symlink target for %v: %v != %v",
					fullPath, linkTarget, spec.linkTarget)
			}
		} else {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("unexpected error reading %v", fullPath)
				return
			}
			if string(content) != fmt.Sprintf("content of %v", spec.path) {
				t.Errorf("unexpected content : %v", fullPath)
			}
		}
	}
}

func TestTar(t *testing.T) {
	os.Mkdir("tmp", 0700)
	sampleDir := path.Join("tmp", "sample")
	samplePaths := makeSamplePaths(sampleDir)
	log.Printf("samplePaths=%v", samplePaths)

	// Create tar file using writeTar()
	tarPath := path.Join("tmp", "archive.tar")
	os.Remove(tarPath)
	tarfile, err := os.Create(tarPath)
	if err != nil {
		log.Fatal("os.Create", err)
	}
	defer tarfile.Close()
	err = writeTar(sampleDir, samplePaths, tarfile, log.Default())
	if err != nil {
		log.Fatalf("writeTar error: %v", err)
	}

	// Extract our tar using standard command
	tarBinary, err := exec.LookPath("tar")
	if err != nil {
		log.Fatal(err)
	}
	extractDir := path.Join("tmp", "extract")
	os.Remove(extractDir)
	os.Mkdir(extractDir, 0700)
	extractCmd := exec.Command(tarBinary, "xf", tarPath, "-C", extractDir)
	if output, err := extractCmd.CombinedOutput(); err != nil {
		log.Fatal("oops ", err, extractCmd.Args, "\n", string(output))
	}

	// Verify
	validateSamplePaths(t, extractDir)
}
