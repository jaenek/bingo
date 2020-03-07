package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var validPath = regexp.MustCompile("^/bingo/(edit|play)/([a-zA-Z0-9]+)$")

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

	path := "bingos/" + p.Name + ".json"
	err = ioutil.WriteFile(path, b, 0666)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"file": path,
	}).Info("Saving to file.")
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

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) error {
	log.WithFields(log.Fields{
		"file": srvPath + tmpl,
	}).Info("Rendering template.")

	t, err := template.ParseFiles(srvPath + tmpl)
	if err != nil {
		return err
	}

	err = t.Execute(w, p)
	if err != nil {
		return err
	}

	return nil
}

func getName(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Name")
	}
	return m[2], nil // The title is the second subexpression.
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/bingo"):]
	if path != "/" {
		http.ServeFile(w, r, srvPath+path)
		log.WithFields(log.Fields{
			"file": path,
		}).Info("Serving file.")
		return
	}

	type Index struct {
		Bingos []string
	}

	files, err := ioutil.ReadDir("bingos/")
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}

	index := &Index{
		Bingos: make([]string, len(files)),
	}

	for i, file := range files {
		fn := file.Name()
		index.Bingos[i] = fn[:len(fn)-len(".json")]
	}

	err = renderTemplate(w, "/index.html", index)
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(w, r)
	if err != nil {
		log.Error(err.Error())
		return
	}

	p, err := read("bingos/" + name + ".json")
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}

	rand.Shuffle(len(p.Phrases), func(i, j int) {
		p.Phrases[i], p.Phrases[j] = p.Phrases[j], p.Phrases[i]
	})

	err = renderTemplate(w, "/play.html", p)
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Phrases: make([]Phrase, 25),
	}

	for i := range p.Phrases {
		p.Phrases[i].ID = uint8(i)
	}

	err := renderTemplate(w, "/add.html", p)
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(w, r)
	if err != nil {
		log.Error(err.Error())
		return
	}

	p, err := read("bingos/" + name + ".json")
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}

	err = renderTemplate(w, "/edit.html", p)
	if err != nil {
		http.NotFound(w, r)
		log.Error(err.Error())
		return
	}
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

	err := p.write()
	if err != nil {
		log.Error(err.Error())
	}

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
