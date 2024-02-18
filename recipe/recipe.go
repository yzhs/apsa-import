package recipe

import (
	"bytes"
	"html/template"
	"log"
	"strings"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/kr/text"
)

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

func (r *Recipe) ToYaml() string {
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
