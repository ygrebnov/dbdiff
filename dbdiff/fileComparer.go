package dbdiff

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ygrebnov/dbdiff/models"
)

// Represents a type capable of comparing two databases as files. Implements [Comparer] interface.
type fileComparer struct{}

// NewFileComparer creates a new fileComparer.
func NewFileComparer() Comparer {
	return &fileComparer{}
}

func (fc *fileComparer) parse(s string, e any) {
	raw, _ := e.(*string)
	// Remove database type prefix from file path
	for _, t := range []models.DatabaseType{sqlite, postgresql} {
		*raw = strings.Replace(s, t.Name()+":", "", 1)
	}
}

func (fc *fileComparer) Compare(_ context.Context, s1 string, s2 string) {
	var fp1, fp2 string
	equal := true // Comparison result holder

	// Parse database identifiers
	fc.parse(s1, &fp1)
	fc.parse(s2, &fp2)

	// Open files
	file1, err := os.OpenFile(s1, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Panicln("Error opening file:", err)
	}
	defer file1.Close()

	file2, err := os.OpenFile(s2, os.O_RDONLY, os.ModePerm)
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
