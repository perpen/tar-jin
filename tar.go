package main

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Writes the bytes for a tar file to the writer
func writeTar(fromDir string, paths []string, w io.Writer, logger *log.Logger) error {
	tw := tar.NewWriter(w)

	var fromDirSlashed string
	if fromDir[len(fromDir)-1] == os.PathSeparator {
		fromDirSlashed = fromDir
	} else {
		fromDirSlashed = fmt.Sprintf("%v/", fromDir)
	}

	// Pushes path into tar if regular file, or dir, or symlink
	push := func(subpath string, info fs.FileInfo) error {
		linkTarget := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			var err error
			linkTarget, err = os.Readlink(subpath)
			if err != nil {
				return err
			}
		} else if !(info.Mode().IsRegular() || info.IsDir()) {
			logger.Printf("ignoring non-regular path: %v", subpath)
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return err
		}
		hdr.Name = strings.TrimPrefix(subpath, fromDirSlashed)
		if info.IsDir() {
			hdr.Name += "/"
		}
		logger.Printf("adding %v", hdr.Name)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if hdr.Size > 0 {
			f, err := os.Open(subpath)
			if err != nil {
				return err
			}
			defer f.Close()
			bf := bufio.NewReader(f)
			if _, err := io.Copy(tw, bf); err != nil {
				return err
			}
		}
		return nil
	}

	for _, p := range paths {
		err := filepath.Walk(filepath.Join(fromDir, p),
			func(subpath string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				push(subpath, info)
				return nil
			})
		if err != nil {
			logger.Printf("error walking the path %q: %v\n", p, err)
			return err
		}
	}
	if err := tw.Close(); err != nil {
		logger.Fatal("tc.Close", err)
	}
	return nil
}
