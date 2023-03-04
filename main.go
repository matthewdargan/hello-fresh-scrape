// Copyright 2023 Matthew Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hello-fresh-scrape extracts recipes from Hello Fresh to JSON output.
//
// Usage:
//
//	hello-fresh-scrape [-l] [-o output] [-p page] [-y]
//
// The -l flag lists available collections to scrape recipes from.
//
// The -o flag specifies the name of a file to write instead of using standard output.
//
// The -p flag specifies the URL of a page to scrape recipes from.
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

const recipeHomePage = "https://www.hellofresh.com/recipes"

var (
	listFlag        = flag.Bool("l", false, "list available collections to scrape recipes from")
	oFlag           = flag.String("o", "", "write output to `file` (default standard output)")
	recipePage      = flag.String("p", "", "URL to scrape recipes from")
	yieldIDsToNames = flag.Bool("y", false, "convert recipe IngredientYield IDs to names")
	output          *bufio.Writer
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hello-fresh-scrape [-l] [-o output] [-p page] [-y]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetPrefix("hello-fresh-scrape: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if *listFlag && *recipePage != "" {
		log.Fatal("cannot use -p with -l")
	}
	if *listFlag && *yieldIDsToNames {
		log.Fatal("cannot use -y with -l")
	}
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
	if *listFlag {
		cs, err := recipe.Collections()
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range cs {
			data = append(data, []byte(c+"\n")...)
		}
	} else {
		if *recipePage == "" {
			*recipePage = "https://www.hellofresh.com/recipes"
		} else if *recipePage != recipeHomePage {
			isValid, err := recipe.IsValidPage(*recipePage)
			if err != nil {
				log.Fatal(err)
			}
			if !isValid {
				log.Fatalf("invalid recipe page: %s", *recipePage)
			}
		}
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
	_, err := output.Write(data)
	if err != nil {
		log.Fatalf("writing recipe output: %v", err)
	}
	err = output.Flush()
	if err != nil {
		log.Fatalf("flushing recipe output: %v", err)
	}
}
