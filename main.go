package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	mm "github.com/edsrzf/mmap-go"
	"github.com/jfcg/sixb"
	"github.com/twotwotwo/sorts"
)

type Lines []string

func (s Lines) Less(i, j int) bool {
	return strings.Compare(s.Key(i), s.Key(j)) == -1
}

func (s Lines) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Lines) Len() int {
	return len(s)
}

func (s Lines) Key(i int) string {
	l := *numChars
	a := s[i]

	if len(a) < l {
		l = len(a)
	}

	return a[:l]
}

func (s Lines) Sort() { sorts.ByString(s) }

var (
	numChars = flag.Int("n", 19, "number of first bytes to use when comparing lines")
)

func main() {
	defaultBufSize := 16 * 1024 * 1024

	sourceFile := flag.String("s", "", "file to sort")
	destFile := flag.String("d", "-", "file to write result to")
	writeBufferSize := flag.Int("write-buffer-size", defaultBufSize,
		"size of write buffer: determines how often data is flushed to disk")
	verbose := flag.Bool("v", false, "verbosity")

	flag.Parse()

	src, err := os.Open(*sourceFile)
	if err != nil {
		log.Fatalf("couldn't open file %v: %v\n", *sourceFile, err)
	}

	defer func() {
		if err := src.Close(); err != nil {
			log.Fatalf("couldn't close file %v: %v\n", *sourceFile, err)
		}
	}()

	now := time.Now()

	if *verbose {
		log.Println("Creating memory-mapped file...")
	}

	m, err := mm.Map(src, mm.RDONLY, 0)

	if err != nil {
		log.Fatalf("couldn't mmap file: %v\n", err)
	}

	numLines := 1

	for i := 0; i < len(m); i++ {
		if m[i] == '\n' {
			numLines++
		}
	}

	if *verbose {
		log.Printf("Created memory-mapped file in %v. Splitting lines...\n", time.Since(now))
		now = time.Now()
	}

	lines := make(Lines, 0, numLines)

	start := 0

	for i := 0; i < len(m); i++ {
		if m[i] == '\n' {
			lines = append(lines, sixb.BtS(m[start:i]))
			start = i + 1
		}
	}

	if len(m) > start {
		lines = append(lines, sixb.BtS(m[start:]))
	}

	if *verbose {
		log.Printf("Splitted file into lines in %v. Sorting...\n", time.Since(now))
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
	}()

	for _, l := range lines {
		_, err := w.Write(sixb.StB(l))
		if err != nil {
			log.Fatalf("couldn't write line to file: %v\n", err)
		}

		if err := w.WriteByte('\n'); err != nil {
			log.Fatalf("couldn't write new line to file: %v\n", err)
		}
	}

	if *verbose {
		log.Printf("Wrote in %v. Closing...\n", time.Since(now))
	}
}
