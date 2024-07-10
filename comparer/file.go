package comparer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	databasePkg "github.com/ygrebnov/dbdiff/entity/database"
)

// file is a type capable of comparing two databases as files.
type file struct {
	f1, f2 string
}

// removePrefixes returns given string without database identifying prefixes.
func removePrefixes(s string) string {
	s = strings.TrimPrefix(s, databasePkg.Sqlite+":")
	s = strings.TrimPrefix(s, databasePkg.Postgresql+":")

	return s
}

// NewFileComparer creates a new fileComparer.
func NewFileComparer(resource1, resource2 string) Comparer {
	return &file{f1: removePrefixes(resource1), f2: removePrefixes(resource2)}
}

func (f *file) Compare(_ context.Context) {
	equal := true // Comparison result holder

	// Open files
	file1, err := os.OpenFile(f.f1, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Panicln("Error opening file:", err)
	}
	defer file1.Close()

	file2, err := os.OpenFile(f.f2, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Panicln("Error opening file:", err)
	}
	defer file2.Close()

	// Create scanners
	f1Scanner := bufio.NewScanner(file1)
	f2Scanner := bufio.NewScanner(file2)

	// Scan two files at the same time
	i := 0
	for {
		i++
		var f1line, f2line string
		if f1 := f1Scanner.Scan(); f1 {
			f1line = f1Scanner.Text()
		}
		if f2 := f2Scanner.Scan(); f2 {
			f2line = f2Scanner.Text()
		}

		// Stop scanning only after both scanners stop
		if f1line == "" && f2line == "" {
			break
		}

		if f1line != f2line {
			if equal {
				fmt.Println("Differences:")
			}
			fmt.Printf("%d: file1: %s, file2: %s\n", i, f1line, f2line)
			equal = false
		}
	}

	if equal {
		fmt.Println("Files are equal")
	}
}
