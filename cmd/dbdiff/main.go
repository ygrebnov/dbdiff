package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ygrebnov/dbdiff/dbdiff"
)

var version, buildTime string

func main() {
	displayHelp := flag.Bool("h", false, "Display help")
	displayVersion := flag.Bool("version", false, "Display version")
	asFiles := flag.Bool("f", false, "Compare databases as files")
	verbose := flag.Bool("v", false, "Level1 verbosity output")
	vverbose := flag.Bool("vv", false, "Level2 verbosity output")
	vvverbose := flag.Bool("vvv", false, "Level3 verbosity output")
	flag.Parse()

	if *displayHelp {
		fmt.Println(help)
		os.Exit(0)
	}

	if *displayVersion {
		fmt.Printf("dbdiff, version: %s, built: %s\n", version, buildTime)
		os.Exit(0)
	}

	if len(flag.Args()) != 2 {
		log.Fatal(usage)
	}

	// take the highest specified verbosity level
	switch {
	case *vvverbose:
		*verbose = false
		*vverbose = false
	case *vverbose:
		*verbose = false
	}

	ctx := context.WithValue(context.Background(), dbdiff.VerboseContextKey, *verbose)
	ctx = context.WithValue(ctx, dbdiff.VVerboseContextKey, *vverbose)
	ctx = context.WithValue(ctx, dbdiff.VVVerboseContextKey, *vvverbose)

	if *asFiles {
		dbdiff.NewFileComparer().Compare(ctx, flag.Args()[0], flag.Args()[1])
	} else {
		dbdiff.NewDatabaseComparer().Compare(ctx, flag.Args()[0], flag.Args()[1])
	}
}
