package main

import (
	filelog "./src/log"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	tt "text/template"
	"time"
)

const (
	servicePrefix     string   = "/service"
	staticDir         http.Dir = http.Dir("public")
	pageTemplate      string   = "page.html"
	storyTemplate     string   = "story.html"
	storyPerPage      int      = 3
	pageStoryTemplate string   = `<div class="blog-post">
            <h2 class="blog-post-title">{{.Title}}</h2>
            <p class="blog-post-meta">{{.CreateTime}}</p>
			<img class="blog-post-img" src="{{.Thumbnail}}">
			{{.Brief}}
			</div>`
)

var (
	storyInfo     []StoryIndex
	staticFileSet        = make(map[string]bool)
	storyInfoFile string = filepath.Join("./data", "storyInfo.json")
	templates            = template.Must(template.ParseFiles(filepath.Join("./tmpl", pageTemplate), filepath.Join("./tmpl", storyTemplate)))
)

type Page struct {
	Content  template.HTML
	Nav      template.HTML
	Featured string
}

type Story struct {
	Title string
	Body  []byte
}

type StoryIndex struct {
	Title      string
	Author     string
	Brief      string
	Thumbnail  string
	CreateTime time.Time
}

func initStoryInfo() {
	b, err := ioutil.ReadFile(storyInfoFile)
	if err != nil {
		filelog.Fatal("unable to init index: ", err)
		return
	}

	err = json.Unmarshal(b, &storyInfo)
	if err != nil {
		filelog.Fatal("unmarshal index failed: ", err)
	}
}

func saveStoryInfo() {
	b, err := json.MarshalIndent(storyInfo, "", "  ")
	if err != nil {
		filelog.Fatal("not be able to save index.")
	}
	ioutil.WriteFile(storyInfoFile, b, 0644)
}

func loadPage(num string) (*Page, error) {
	total := len(storyInfo)
	pages := int(math.Ceil(float64(total) / float64(storyPerPage)))
	n, err := strconv.Atoi(num)
	if err != nil {
		return nil, err
	}
	if n > pages || n < 1 {
		return nil, errors.New("page num is wrong")
	}
	begin := storyPerPage * (n - 1)
	t := tt.Must(tt.New("pageStoryTemplate").Parse(pageStoryTemplate))
	var post bytes.Buffer
	for i := begin; (i < begin+storyPerPage) && (i < total); i++ {
		t.Execute(&post, storyInfo[i])
	}

	var nav bytes.Buffer
	nav.WriteString(`<nav><ul class="pager">`)
	if n > 1 {
		nav.WriteString(fmt.Sprintf(`<li><a href="/service/page/%d">前一页</a></li>`, n-1))
	}
	if n < pages {
		nav.WriteString(fmt.Sprintf(`<li><a href="/service/page/%d">后一页</a></li>`, n+1))
	}
	nav.WriteString("</ul></nav>")

	return &Page{Content: template.HTML(post.String()), Nav: template.HTML(nav.String()), Featured: ""}, nil
}

func loadStory(title string) (*Story, error) {
	filename := "data" + string(os.PathSeparator) + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Story{Title: title, Body: body}, nil
}

func pageViewHandler(w http.ResponseWriter, r *http.Request) {
	num := r.URL.Path[len("/service/page/"):]
	p, err := loadPage(num)
	if err != nil {
		filelog.Error("load page error: ", err)
		http.NotFound(w, r)
		return
	}
	filelog.Debug("start to render page")
	renderTemplate(w, pageTemplate, p)
}

func storyViewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/service/story/"):]
	a, err := loadStory(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, storyTemplate, a)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func clean() {
	filelog.Info("log clean, exiting...")
	saveStoryInfo()
	filelog.Close()
	os.Exit(0)
}

func main() {
	flag.Parse()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sc
		filelog.Info("signal recieved: %v", s)
		clean()
	}()

	initStoryInfo()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/service/page/", pageViewHandler)
	http.HandleFunc("/service/story/", storyViewHandler)

	fmt.Println("start service at :8180")
	go func() {
		http.ListenAndServe(":8180", nil)
	}()
	<-sc
}
