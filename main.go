// Copyright 2023 Matthew Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hello-fresh-scrape extracts recipes from Hello Fresh to JSON output.
//
// Usage:
//
//	hello-fresh-scrape [-l list] [-o output] [-y]
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
	listCollections = flag.Bool("l", false, "list available collections to scrape recipes from")
	oFlag           = flag.String("o", "", "write output to `file` (default standard output)")
	recipePage      = flag.String("p", "https://www.hellofresh.com/recipes", "page to scrape recipes from")
	yieldIDsToNames = flag.Bool("y", false, "convert recipe IngredientYield IDs to names")
	output          *bufio.Writer
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hello-fresh-scrape [-l list] [-o output] [-y]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetPrefix("hello-fresh-scrape: ")
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
	var data []byte
	cs, err := recipe.Collections()
	if err != nil {
		log.Fatal(err)
	}
	if *listCollections {
		for _, c := range cs {
			data = append(data, []byte(c+"\n")...)
		}
	} else {
		//for _, c := range cs {
		//	// TODO: Find recipePage in collections.
		//}
		rs, err := recipe.ScrapeRecipes(*recipePage)
		if err != nil {
			log.Fatal(err)
		}
		if *yieldIDsToNames {
			err = rs.YieldIDsToNames()
			if err != nil {
				log.Fatal(err)
			}
		}
		data, err = json.MarshalIndent(rs, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = output.Write(data)
	if err != nil {
		log.Fatalf("writing recipe output: %v", err)
	}
	err = output.Flush()
	if err != nil {
		log.Fatalf("flushing recipe output: %v", err)
	}
}
