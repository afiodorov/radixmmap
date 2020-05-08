package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	_ "net/http/pprof"

	mm "github.com/edsrzf/mmap-go"
	"github.com/twotwotwo/sorts"
)

type Slice struct {
	start int
	end   int
}

type Lines struct {
	data   []byte
	slices []Slice
}

func (l Lines) Line(i int) []byte {
	return l.data[l.slices[i].start:l.slices[i].end]
}

func (s Lines) Less(i, j int) bool {
	limI, limJ := *numChars, *numChars

	a := s.Line(i)
	b := s.Line(j)

	if len(a) < limI {
		limI = len(a)
	}

	if len(b) < limJ {
		limJ = len(b)
	}

	return bytes.Compare(a[:limI], b[:limJ]) == -1
}

func (s Lines) Swap(i, j int) {
	s.slices[i], s.slices[j] = s.slices[j], s.slices[i]
}

func (s Lines) Len() int {
	return len(s.slices)
}

func (s Lines) Key(i int) []byte {
	l := *numChars
	a := s.Line(i)

	if len(a) < l {
		l = len(a)
	}

	return a[:l]
}

func (s Lines) Sort() { sorts.ByBytes(s) }

var (
	numChars = flag.Int("n", 19, "number of first bytes to use when comparing lines")
)

func main() {
	defaultBufSize := 32 * 1024 * 1024
	newLine := byte(10)

	sourceFile := flag.String("s", "", "file to sort")
	destFile := flag.String("d", "-", "file to write result to")
	writeBufferSize := flag.Int("write-buffer-size", defaultBufSize,
		"size of write buffer: determines how often data is flushed to disk")

	flag.Parse()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	src, err := os.Open(*sourceFile)
	if err != nil {
		log.Fatalf("couldn't open file %v: %v\n", *sourceFile, err)
	}

	defer func() {
		if err := src.Close(); err != nil {
			log.Fatalf("couldn't close file %v: %v\n", *sourceFile, err)
		}
	}()

	m, err := mm.Map(src, mm.RDONLY, 0)
	if err != nil {
		log.Fatalf("couldn't mmap file: %v\n", err)
	}

	numLines := 1

	for i := 0; i < len(m); i++ {
		if m[i] == newLine {
			numLines++
		}
	}

	lines := Lines{data: m, slices: make([]Slice, 0, numLines)}

	start := 0

	for i := 0; i < len(m); i++ {
		if m[i] == newLine {
			lines.slices = append(lines.slices, Slice{start: start, end: i})
			start = i + 1
		}
	}

	if int(len(m)) > start {
		lines.slices = append(lines.slices, Slice{start: start, end: len(m)})
	}

	lines.Sort()

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
			log.Fatalf("couldn't flush file: %v\n", err)
		}
	}()

	for _, l := range lines.slices {
		_, err := w.Write(lines.data[l.start:l.end])
		if err != nil {
			log.Fatalf("couldn't write line to file: %v\n", err)
		}

		if err := w.WriteByte(newLine); err != nil {
			log.Fatalf("couldn't write new line to file: %v\n", err)
		}
	}
}
