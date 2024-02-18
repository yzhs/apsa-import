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

	sprig "github.com/Masterminds/sprig/v3"
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

	CookTime  string `json:"cookTime"`
	PrepTime  string `json:"prepTime"`
	TotalTime string `json:"totalTime"`

	Nutrition struct {
		servingSize  string `json:"servingSize"`
		calories     string `json:"calories"`
		protein      string `json:"proteinContent"`
		fat          string `json:"fatContent"`
		carbohydrate string `json:"carbohydrateContent"`
	} `json:"nutrition"`

	Ingredients  []string `json:"recipeIngredient"`
	Instructions string   `json:"recipeInstructions"`
}

func (r *Recipe) toYaml() string {
	const tmplString = `title: {{.Title}}
source: {{.Source}}
tags:
  {{- range .Tags }}
  - {{ . }}
  {{- end }}
yield: {{.Yield}}
time:
{{- with .CookTime }}
  cooking: {{ . }}
{{- end }}
{{- with .PrepTime }}
  preparation: {{ . }}
{{- end }}
{{- with .TotalTime }}
  total: {{ . }}
{{- end }}

steps:
- ingredients:
  {{- range .Ingredients }}
  - {{ . }}
  {{- end }}
  instructions: |
{{.Instructions | indent 4}}
`

	tmpl := template.Must(template.New("recipe.yaml").Funcs(sprig.TxtFuncMap()).Parse(tmplString))
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, r)
	if err != nil {
		log.Panic(err)
	}
	return buf.String()
}

func (recipe *Recipe) Polish() {
	for i, ingredient := range recipe.Ingredients {
		recipe.Ingredients[i] = strings.TrimSpace(ingredient)
	}
	tmp := ""
	for _, str := range strings.Split(recipe.Instructions, "\n") {
		tmp += text.Wrap(str, 80) + "\n"
	}
	recipe.Instructions = tmp
}

func generateRecipe(url string) string {
	url = strings.TrimSpace(url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Panic(err)
	}

	recipe := ConvertRecipe(url, doc)
	recipe.Polish()
	return recipe.toYaml()
}

func handleURL(url string) {
	if !strings.HasPrefix(url, "http") {
		return
	}
	id := uuid.NewV4().String()
	f, err := os.OpenFile(homeDir+"/.apsa/library/"+id+".yaml", os.O_CREATE|os.O_WRONLY, 0644)
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
