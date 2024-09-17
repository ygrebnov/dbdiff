package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ygrebnov/dbdiff/comparer"
)

var version, buildTime string

func main() {
	displayHelp := flag.Bool("h", false, "Display help")
	displayVersion := flag.Bool("version", false, "Display version")
	asFiles := flag.Bool("f", false, "Compare databases as files")
	verbosity1 := flag.Bool("v", false, "Level1 verbosity output")
	verbosity2 := flag.Bool("vv", false, "Level2 verbosity output")
	verbosity3 := flag.Bool("vvv", false, "Level3 verbosity output")
	flag.Parse()

	if *displayHelp {
		fmt.Println(help)
		os.Exit(0)
	}

	if *displayVersion {
		fmt.Printf("dbdiff, version: %s, built: %s\n", version, buildTime)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) != 2 {
		log.Fatal(usage)
	}

	var (
		c         comparer.Comparer
		verbosity int
	)

	switch {
	case *verbosity1:
		verbosity = 1
	case *verbosity2:
		verbosity = 2
	case *verbosity3:
		verbosity = 3
	}

	if *asFiles {
		c = comparer.NewFileComparer(args[0], args[1])
	} else {
		c = comparer.NewDatabaseComparer(args[0], args[1], verbosity)
	}

	c.Compare(context.Background())
}
