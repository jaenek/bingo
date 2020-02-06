package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
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

func (p *Page) write() error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("bingos/"+p.Name+".json", b, 0666)
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

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getPath(fullpath string) string {
	return fullpath[len("/bingo"):]
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	path := getPath(r.URL.Path)
	if path != "/" {
		http.ServeFile(w, r, srvPath+path)
		return
	}

	type Index struct {
		Bingos []string
	}

	files, err := ioutil.ReadDir("bingos/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	index := &Index{
		Bingos: make([]string, len(files)),
	}

	for i, file := range files {
		fn := file.Name()
		index.Bingos[i] = fn[:len(fn)-5]
	}

	renderTemplate(w, srvPath+"/index.html", index)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	name := getPath(r.URL.Path)[len("/play/"):]
	if len(name) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	p, err := read("bingos/" + name + ".json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rand.Shuffle(len(p.Phrases), func(i, j int) {
		p.Phrases[i], p.Phrases[j] = p.Phrases[j], p.Phrases[i]
	})

	renderTemplate(w, srvPath+"/play.html", p)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Phrases: make([]Phrase, 25),
	}

	for i := range p.Phrases {
		p.Phrases[i].ID = uint8(i)
	}

	renderTemplate(w, srvPath+"/add.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	name := getPath(r.URL.Path)[len("/edit/"):]
	if len(name) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	p, err := read("bingos/" + name + ".json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, srvPath+"/edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Name:    r.FormValue("name"),
		Title:   r.FormValue("title"),
		Phrases: make([]Phrase, 25),
	}

	for i := range p.Phrases {
		p.Phrases[i].ID = uint8(i)
		p.Phrases[i].Phrase = r.FormValue(strconv.Itoa(i))
	}

	p.write()

	http.Redirect(w, r, "/play/"+p.Name, http.StatusFound)
}
func main() {
	http.HandleFunc("/bingo/play/", playHandler)
	http.HandleFunc("/bingo/add", addHandler)
	http.HandleFunc("/bingo/edit/", editHandler)
	http.HandleFunc("/bingo/save", saveHandler)
	http.HandleFunc("/bingo/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
