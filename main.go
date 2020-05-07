package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

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
	limI, limJ := 20, 20

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
	l := 20
	a := s.Line(i)

	if len(a) < l {
		l = len(a)
	}

	return a[:l]
}

func (s Lines) Sort() { sorts.ByBytes(s) }

func main() {
	fileName := flag.String("s", "", "file to sort")
	flag.Parse()

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("%v\n", err)
		}
	}()

	m, err := mm.Map(file, mm.RDONLY, 0)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	defer func() {
		if err := m.Flush(); err != nil {
			log.Fatalf("%v\n", err)
		}
	}()

	lines := Lines{data: m, slices: make([]Slice, 0)}

	start := 0

	for i := 0; i < len(m); i++ {
		if m[i] == byte(10) {
			lines.slices = append(lines.slices, Slice{start: start, end: i})
			start = i + 1
		}
	}

	if len(m) > start {
		lines.slices = append(lines.slices, Slice{start: start, end: len(m)})
	}

	lines.Sort()

	for _, l := range lines.slices {
		fmt.Printf("%s\n", lines.data[l.start:l.end])
	}
}
