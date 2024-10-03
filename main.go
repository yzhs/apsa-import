package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/PuerkitoBio/goquery"
	uuid "github.com/satori/go.uuid"

	"github.com/yzhs/apsa-import/chefkoch"
	"github.com/yzhs/apsa-import/recipe"
)

var homeDir string

func usage() {
	fmt.Println("Usage: apsa-import [URL]...")
}

func convertRecipe(url string, doc *goquery.Document) recipe.Recipe {
	return chefkoch.ConvertRecipe(url, doc)
}

func generateRecipe(url string) string {
	url = strings.TrimSpace(url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Panic(err)
	}

	recipe := convertRecipe(url, doc)
	recipe.Polish()
	return recipe.ToYaml()
}

func getConfirmation() bool {
	fmt.Print("Looks good? [Y/n] ")

	reader := bufio.NewScanner(os.Stdin)
	response := reader.Text()

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes" || response == ""
}

func handleURL(url string) {
	if !strings.HasPrefix(url, "http") {
		return
	}
	recipe := generateRecipe(url)
	fmt.Printf("%v\n\n", recipe)
	good := getConfirmation()

	if !good {
		fmt.Println("Discarding recipe")
		return
	}

	id := uuid.NewV4().String()
	f, err := os.OpenFile(homeDir+"/.apsa/library/"+id+".yaml", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(recipe)
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
