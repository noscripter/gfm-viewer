package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pocke/gfm-viewer/env"
	"github.com/yosssi/ace"
)

type Server struct {
	pages  map[string]string
	mu     *sync.RWMutex
	router *mux.Router
}

func NewServer() *Server {
	r := mux.NewRouter()

	s := &Server{
		pages:  make(map[string]string),
		mu:     &sync.RWMutex{},
		router: r,
	}

	go func() {
		http.Handle("/", r)
		// TODO: port
		http.ListenAndServe(":1124", nil)
	}()

	// TODO: index
	r.HandleFunc("/files/{path}", s.ServeFile)
	return s
}

func (s *Server) Add(path, html string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pages[path] = html
}

// Update same as Add.
func (s *Server) Update(path, html string) {
	s.Add(path, html)
}

func (s *Server) Get(path string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	html, ok := s.pages[path]
	return html, ok
}

func (s *Server) ServeFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	html, ok := s.Get(path)
	if !ok {
		http.Error(w, fmt.Sprintf("%s page not found", path), http.StatusNotFound)
		return
	}
	w.Write([]byte(html))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := ace.Load("assets/index", "", &ace.Options{
		DynamicReload: env.DEBUG,
		Asset:         Asset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
