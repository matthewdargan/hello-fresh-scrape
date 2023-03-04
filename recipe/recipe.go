// Copyright 2023 Matthew Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package recipe extracts recipes from Hello Fresh.
package recipe

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	XMLName    xml.Name `xml:"url"`
	LOC        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod"`
	ChangeFreq string   `xml:"changefreq"`
	Priority   float64  `xml:"priority"`
}

func Collections() ([]string, error) {
	resp, err := http.Get("https://www.hellofresh.com/sitemap_recipe_collections.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var urlset URLSet
	err = xml.Unmarshal(b, &urlset)
	if err != nil {
		return nil, err
	}
	var collection []string
	for _, url := range urlset.URLs {
		collection = append(collection, url.LOC)
	}
	return collection, nil
}

type payload struct {
	Props struct {
		PageProps struct {
			SSRPayload struct {
				DehydratedState struct {
					Queries []struct {
						State struct {
							Data json.RawMessage
						}
					}
				}
			}
		}
	}
}

type data struct {
	Items []Recipe
}

type Recipe struct {
	ID                  string
	Country             string
	Name                string
	SeoName             string
	Category            Category
	Slug                string
	Headline            string
	Description         string
	DescriptionHTML     string
	DescriptionMarkdown string
	SeoDescription      string
	Comment             string
	Difficulty          int
	PrepTime            string
	TotalTime           string
	ServingSize         int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Link                string
	ImageLink           string
	ImagePath           string
	CardLink            string
	VideoLink           string
	Nutrition           []Nutrition
	Ingredients         []Ingredient
	Allergens           []Allergen
	Utensils            []Utensil
	Tags                []Tag
	Cuisines            []Cuisine
	Yields              []Yield
}

type Recipes []Recipe

type Category struct {
	ID       string
	Type     string
	Name     string
	Slug     string
	IconLink string
	IconPath string
	Usage    int
}

type Nutrition struct {
	Type   string
	Name   string
	Amount float64
	Unit   string
}

type Ingredient struct {
	Country           string
	ID                string
	Type              string
	Name              string
	Slug              string
	Description       string
	InternalName      string
	Shipped           bool
	ImageLink         string
	ImagePath         string
	Usage             int
	HasDuplicatedName bool
	Allergens         []string
	Family            IngredientFamily
	UUID              string
}

type IngredientFamily struct {
	ID             string
	Type           string
	Description    string
	Priority       int
	IconLink       string
	IconPath       string
	UsageByCountry map[string]int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UUID           string
	Name           string
	Slug           string
}

type Allergen struct {
	ID               string
	Type             string
	Description      string
	TracesOf         bool
	TriggersTracesOf bool
	IconLink         string
	IconPath         string
	Usage            int
	Name             string
	Slug             string
}

type Utensil struct {
	ID   string
	Type string
	Name string
}

type Tag struct {
	ID                       string
	Type                     string
	IconLink                 string
	IconPath                 string
	NumberOfRecipes          int
	NumberOfRecipesByCountry map[string]int
	ColorHandle              string
	Preferences              []string
	Name                     string
	Slug                     string
	DisplayLabel             bool
}

type Cuisine struct {
	ID       string
	Type     string
	IconLink string
	IconPath string
	Usage    int
	Name     string
	Slug     string
}

type Yield struct {
	Yields      int
	Ingredients []IngredientYield
}

type IngredientYield struct {
	ID     string
	Amount float64
	Unit   string
}

// ScrapeRecipes scrapes recipes from the JSON payload on the
// Hello Fresh website.
func ScrapeRecipes(page string) (Recipes, error) {
	resp, err := http.Get(page)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := parseRecipeProps(resp.Body)
	if err != nil {
		return nil, err
	}
	var p payload
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}
	var rs Recipes
	for _, q := range p.Props.PageProps.SSRPayload.DehydratedState.Queries {
		// Recipes only occur when Data is a JSON object
		if q.State.Data[0] == '{' {
			var d data
			err = json.Unmarshal(q.State.Data, &d)
			if err != nil {
				return nil, err
			}
			if len(d.Items) > 0 {
				rs = append(rs, d.Items...)
			}
		}
	}
	return rs, nil
}

func parseRecipeProps(r io.Reader) ([]byte, error) {
	z := html.NewTokenizer(r)
	isRecipeProps := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, z.Err()
		case html.TextToken:
			if isRecipeProps {
				return z.Text(), nil
			}
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if string(tn) == "script" && hasAttr {
				k, v, moreAttr := z.TagAttr()
				if string(k) == "id" && string(v) == "__NEXT_DATA__" && moreAttr {
					k, v, _ = z.TagAttr()
					if string(k) == "type" && string(v) == "application/json" {
						isRecipeProps = true
					}
				}
			}
		}
	}
	return nil, errors.New("recipe props data not found")
}

// YieldIDsToNames converts recipe IngredientYield IDs to their
// respective names.
func (rs Recipes) YieldIDsToNames() error {
	for _, r := range rs {
		for _, ys := range r.Yields {
			for i, ingred := range ys.Ingredients {
				name, err := ingredientName(ingred.ID, r.Ingredients)
				if err != nil {
					return err
				}
				ys.Ingredients[i].ID = name
			}
		}
	}
	return nil
}

func ingredientName(id string, ingreds []Ingredient) (string, error) {
	for _, ingred := range ingreds {
		if id == ingred.ID {
			return ingred.Name, nil
		}
	}
	return "", errors.New(fmt.Sprintf("id %s not found in ingredients list", id))
}
