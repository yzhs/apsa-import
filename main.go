package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/kr/text"
	uuid "github.com/satori/go.uuid"
)

var homeDir string

func usage() {
	fmt.Println("Usage: apsa-import-chefkoch.de [URL]...")
}

type Recipe struct {
	Title  string `json:"name"`
	Source string
	Tags   []string `json:"keywords"`
	Yield  string   `json:"recipeYield"`

	CookTime string `json:"cookTime"`
	PrepTime string `json:"prepTime"`

	Ingredients  []string `json:"recipeIngredient"`
	Instructions string   `json:"recipeInstructions"`
}

func (r *Recipe) toApsa() string {
	const tmplString = `# {{.Title}}
Quelle: {{.Source}}
Tags: {{range .Tags}}{{.}}, {{end}}
Portionen: {{.Yield}}
Kochzeit: {{.CookTime}}
Zubereitungszeit: {{.PrepTime}}

Zutaten:
{{range .Ingredients}}* {{.}}
{{end}}
{{.Instructions}}

`

	tmpl := template.Must(template.New("Recipe").Parse(tmplString))
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, r)
	if err != nil {
		log.Panic(err)
	}
	return buf.String()
}

func generateRecipe(url string) string {
	// Read page
	url = strings.TrimSpace(url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Panic(err)
	}

	// Extract data
	var recipe Recipe
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

				err = json.Unmarshal([]byte(tmp), &recipe)
				if err != nil {
					log.Panic(err)
				}

				// Polish recipe
				recipe.Source = url
				for i, ingredient := range recipe.Ingredients {
					recipe.Ingredients[i] = strings.TrimSpace(ingredient)
				}
				recipe.CookTime = strings.Replace(strings.TrimPrefix(recipe.CookTime, "PT"), "M", " min", 1)
				recipe.PrepTime = strings.Replace(strings.TrimPrefix(recipe.PrepTime, "PT"), "M", " min", 1)
				tmp = ""
				for _, str := range strings.Split(recipe.Instructions, "\n") {
					tmp += text.Wrap(str, 80) + "\n"
				}
				recipe.Instructions = tmp
			}
		},
	)

	// Generate output
	return recipe.toApsa()
}

func handleURL(url string) {
	if !strings.HasPrefix(url, "http") {
		return
	}
	id := uuid.NewV4().String()
	f, err := os.OpenFile(homeDir+"/.apsa/library/"+id+".md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(generateRecipe(url))
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		usage()
		os.Exit(0)
	}
	u, err := user.Current()
	if err != nil {
		log.Panic(err)
	}
	homeDir = u.HomeDir

	if len(os.Args) == 1 {
		// Read URLs from stdin
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Panic(err)
		}
		for _, url := range strings.Split(string(bytes), "\n") {
			handleURL(url)
		}
	} else {
		i := 0
		for _, url := range os.Args[1:] {
			handleURL(url)
			i += 1
		}
		fmt.Println("Sucessfully imported", i, "recipes.")
	}

	err = exec.Command("apsa", "-i").Run()
	if err != nil {
		panic(err)
	}
}
