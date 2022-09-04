package main

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %v ARCHIVE PATH...\n  eg %v archive.tar.gz blah1 /blah2\n",
			os.Args[0], os.Args[0])
		os.Exit(2)
	}
	archive := os.Args[1]
	paths := os.Args[2:]

	tarfile, err := os.Create(archive)
	if err != nil {
		log.Fatal("os.Create", err)
	}
	defer tarfile.Close()

	gzipWriter := gzip.NewWriter(tarfile)
	defer gzipWriter.Close()

	err = writeTar(".", paths, gzipWriter, log.Default())
	if err != nil {
		log.Fatalf("writeTar error: %v", err)
	}
}
