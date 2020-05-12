package main

import (
	"strings"

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
