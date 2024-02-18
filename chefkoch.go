package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yzhs/apsa-import-chefkoch.de/recipe"
)

func convertRecipe(url string, doc *goquery.Document) recipe.Recipe {
	var recipe recipe.Recipe
	doc.Find(`script[type="application/ld+json"]`).Each(
		func(i int, s *goquery.Selection) {
			txt := strings.Replace(strings.TrimSpace(s.Text()), " , ", ", ", -1)
			if strings.Contains(txt, "\"@type\": \"Recipe\"") {
				tmp := ""
				inString := false
				for _, chr := range txt {
					if chr == '\n' && inString {
						tmp += "\\n"
						continue
					}

					tmp += string(chr)
					if chr == '"' {
						inString = !inString
					}
				}
				tmp = strings.Replace(tmp, "&szlig;", "ß", -1)
				tmp = strings.Replace(tmp, "&amp;", "&", -1)
				tmp = strings.Replace(tmp, "\t", " ", -1)

				err := json.Unmarshal([]byte(tmp), &recipe)
				if err != nil {
					log.Panic(err)
				}

				recipe.Source = url
			}
		},
	)

	return recipe
}
