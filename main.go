package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/jfcg/sixb"
)

var (
	numChars = flag.Int("n", 19, "number of first bytes to use when comparing lines")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... FILE...\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	defaultBufSize := 16 * 1024 * 1024

	destFile := flag.String("d", "-", "file to write result to")
	writeBufferSize := flag.Int("write-buffer-size", defaultBufSize,
		"size of write buffer: determines how often data is flushed to disk")
	verbose := flag.Bool("v", false, "verbosity")
	skipHeader := flag.Bool("skip-header", false, "skip header of each file")

	flag.Parse()

	fileNames := flag.Args()
	if len(fileNames) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	fileMetas := make([]FileMeta, len(fileNames))
	maxGoroutines := runtime.NumCPU()
	guard := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	now := time.Now()
	if *verbose {
		log.Printf("Creating %v memory-mapped files...\n", len(fileNames))
	}
	for i, file := range fileNames {
		guard <- struct{}{}
		src, err := os.Open(file)
		if err != nil {
			log.Fatalf("couldn't open file %v: %v\n", file, err)
		}

		wg.Add(1)
		go func(src *os.File, i int) {
			defer wg.Done()
			memoryMapFile(src, fileMetas, i, *skipHeader)
			<-guard
		}(src, i)

		defer func(src *os.File) {
			if err := src.Close(); err != nil {
				log.Fatalf("couldn't close file %v: %v\n", src.Name(), err)
			}
		}(src)
	}

	if *verbose {
		log.Printf("Created memory-mapped files in %v. Allocating memory to hold lines...\n", time.Since(now))
		now = time.Now()
	}

	wg.Wait()

	var numLines int
	for _, fileMeta := range fileMetas {
		numLines += fileMeta.NumLines
	}

	lines := make(Lines, numLines)

	if *verbose {
		log.Printf("Allocated memory for %v lines in %v. Splitting into lines...\n", numLines, time.Since(now))
		now = time.Now()
	}

	var numLinesBefore int
	for _, fileMeta := range fileMetas {
		guard <- struct{}{}

		wg.Add(1)
		go func(content []byte, numLinesBefore int) {
			defer wg.Done()
			lineSplitter(lines, content, numLinesBefore)
			<-guard
		}(fileMeta.Content, numLinesBefore)

		numLinesBefore += fileMeta.NumLines
	}

	wg.Wait()

	if *verbose {
		log.Printf("Splitted files into lines in %v. Sorting...\n", time.Since(now))
		now = time.Now()
	}

	lines.Sort()

	if *verbose {
		log.Printf("Sorted in %v. Writing...\n", time.Since(now))
		now = time.Now()
	}

	var dst io.Writer = os.Stdout

	if *destFile != "-" {
		dstFile, err := os.Create(*destFile)
		if err != nil {
			log.Fatalf("couldn't create file %v: %v\n", *destFile, err)
		}

		defer func() {
			if err := dstFile.Close(); err != nil {
				log.Fatalf("couldn't close file %v, %v\n", *destFile, err)
			}
		}()

		dst = dstFile
	}

	w := bufio.NewWriterSize(dst, *writeBufferSize)

	defer func() {
		if err := w.Flush(); err != nil {
			log.Fatalf("couldn't flush writer: %v\n", err)
		}

		if *verbose {
			log.Printf("Wrote in %v. Closing...\n", time.Since(now))
		}
	}()

	for _, l := range lines {
		_, err := w.Write(sixb.StB(l))
		if err != nil {
			log.Fatalf("couldn't write line to file: %v\n", err)
		}
	}
}
