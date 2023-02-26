// Copyright 2023 Matthew Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hello-fresh-scrape extracts recipes from Hello Fresh to JSON output.
//
// Usage:
//
//	hello-fresh-scrape [-o output] [-y]
//
// The -o flag specifies the name of a file to write instead of using standard output.
//
// The -y flag converts recipe IngredientYield IDs to names.
package main // import "github.com/matthewdargan/hello-fresh-scrape"

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/matthewdargan/hello-fresh-scrape/recipe"
)

var (
	oFlag           = flag.String("o", "", "write output to `file` (default standard output)")
	yieldIDsToNames = flag.Bool("y", false, "convert recipe IngredientYield IDs to names")
	output          *bufio.Writer
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hello-fresh-scrape [-o output] [-y]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetPrefix("hello-fresh-extract: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	outfile := os.Stdout
	if *oFlag != "" {
		f, err := os.Create(*oFlag)
		if err != nil {
			log.Fatal(err)
		}
		outfile = f
	}
	output = bufio.NewWriter(outfile)
	rs, err := recipe.ScrapeRecipes()
	if err != nil {
		log.Fatal(err)
	}
	if *yieldIDsToNames {
		err = rs.YieldIDsToNames()
		if err != nil {
			log.Fatal(err)
		}
	}
	d, err := json.MarshalIndent(rs, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	_, err = output.Write(d)
	if err != nil {
		log.Fatalf("writing recipe output: %v", err)
	}
	err = output.Flush()
	if err != nil {
		log.Fatalf("flushing recipe output: %v", err)
	}
}
