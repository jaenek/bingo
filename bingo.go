package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

const (
	srvPath string = "./public"
)

type Phrase struct {
	ID     uint8  `json: "id"`
	Phrase string `json: "phrase"`
}

type Page struct {
	Name    string   `json: "name"`
	Title   string   `json: "title"`
	Phrases []Phrase `json: "phrases"`
}

func (p *Page) write(path string) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0666)
	if err != nil {
		return err
	}

	return nil
}

func read(path string) (*Page, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &Page{}
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil
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

	p, err := read("bingos/lioli1.json")
	if err != nil {
		log.Println(err)
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
