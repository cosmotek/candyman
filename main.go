package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Targets struct {
	Data []string `json:"targets"`
}

type Recipe struct {
	Title        string
	Description  string
	PreviewImage string
	Ingredients  []string
	Instructions []string
	PrepTime     string
	CookTime     string
	Categories   []string
	Cuisine      string
	Servings     string
	Notes        []string
	Source       string
}

func main() {
	targetFile, err := os.Open("targets.json")
	if err != nil {
		panic(err)
	}

	targets := Targets{}
	err = json.NewDecoder(targetFile).Decode(&targets)
	if err != nil {
		panic(err)
	}

	browser := createVisibleBrowser()

	for _, target := range targets.Data {
		siteData, err := ScrapeWebsite(browser, target)
		if err != nil {
			panic(err)
		}

		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "\t")
		err = e.Encode(siteData)
		if err != nil {
			panic(err)
		}
	}

	// time.Sleep(time.Minute * 100)
}

func createVisibleBrowser() *rod.Browser {
	browserStartOpts := launcher.New().
		Headless(true).
		Devtools(false)

	return rod.New().
		ControlURL(browserStartOpts.MustLaunch()).
		Trace(false).
		// SlowMotion(time.Second * 2).
		MustConnect()
}

var ingredientsKeywords = []string{
	"ingredients",
	"ingredient",
}

var instructionsKeywords = []string{
	"instructions",
	"instruction",
	"direction",
	"directions",
	"steps",
}

func ScrapeWebsite(browser *rod.Browser, target string) (Recipe, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: target})
	if err != nil {
		return Recipe{}, err
	}
	defer page.Close()

	err = page.WaitLoad()
	if err != nil {
		return Recipe{}, err
	}

	elems := page.MustElements("h1, h2, h3")
	ingredients := []string{}
	instructions := []string{}

	for _, elem := range elems {
		text, err := elem.Text()
		if err != nil {
			return Recipe{}, err
		}

		for _, keyword := range ingredientsKeywords {
			if strings.Contains(strings.ToLower(strings.TrimSpace(text)), keyword) {
				parent, err := elem.Parent()
				if err != nil {
					return Recipe{}, err
				}

				listItems := parent.MustElements("li")
				for _, item := range listItems {
					ingredients = append(ingredients, item.MustText())
				}
			}
		}

		for _, keyword := range instructionsKeywords {
			if strings.Contains(strings.ToLower(strings.TrimSpace(text)), keyword) {
				parent, err := elem.Parent()
				if err != nil {
					return Recipe{}, err
				}

				listItems := parent.MustElements("li")
				for _, item := range listItems {
					instructions = append(instructions, item.MustText())
				}
			}
		}
	}

	// TODO, if og meta tags not found, try using twitter tags
	metaTitle := page.MustElement("[property='og:title']")
	metaDescription := page.MustElement("[property='og:description']")
	metaImage := page.MustElement("[property='og:image']")

	return Recipe{
		Title:        *metaTitle.MustAttribute("content"),
		Description:  *metaDescription.MustAttribute("content"),
		PreviewImage: *metaImage.MustAttribute("content"),
		Ingredients:  ingredients,
		Instructions: instructions,
		Source:       target,
	}, nil
}
