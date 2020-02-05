package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
)

const (
	srvPath = "./public"
)

type Phrase struct {
	ID     uint8
	Phrase string
}

type Page struct {
	Title   string
	Phrases []Phrase
}

var data = []string{
	"inspektor",
	"azbest",
	"zwiększenie budżetu",
	"komplikacje remontowe",
	"facetka chce przeprowadzki",
	"facet chce przeprowadzki",
	"Jillian albo projektantki mają brzuch",
	"Jillian robi coś specjalnego",
	"Todd szuka poza dzielnicą",
	"Todd pokazuje dom do wykończenia",
	"\"We are going to list it\"",
	"\"We are going to love it\"",
	"\"Jillian świetnie się spisała\"",
	"\"Todd musi się bardziej postarać\"",
	"\"duża otwarta przestrzeń\"",
	"\"spacious\"",
	"\"nice views\"",
	"\"hardwood flooring\"",
	"<++>",
	"<++>",
	"<++>",
	"<++>",
	"<++>",
	"<++>",
	"<++>",
}

var p = &Page{
	Title:   "Love it or List it bingo",
	Phrases: make([]Phrase, 25),
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.ServeFile(w, r, srvPath+r.URL.Path)
		return
	}

	for i, phrase := range data {
		p.Phrases[i] = Phrase{
			ID:     uint8(i),
			Phrase: phrase,
		}
	}

	rand.Shuffle(len(p.Phrases), func(i, j int) {
		p.Phrases[i], p.Phrases[j] = p.Phrases[j], p.Phrases[i]
	})

	renderTemplate(w, srvPath+"/index.html", p)
}

func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
