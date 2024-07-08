package router

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

type URLRouter struct {
	chi.Router
	Service Service
}

type Service interface {
	Create(value string) (ID string)
	Get(ID string) (string, bool)
}

func NewURLRouter(s Service) *URLRouter {
	r := URLRouter{chi.NewRouter(), s}

	r.Get("/{id}", r.handleGet)
	r.Post("/", r.handlePost)

	return &r
}

func (router *URLRouter) handlePost(w http.ResponseWriter, r *http.Request) {
	linkBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Couldn't read body", http.StatusBadRequest)
		return
	}
	link := string(linkBytes)
	if len(link) == 0 {
		http.Error(w, "Link string can't be empty", http.StatusBadRequest)
		return
	}

	fmt.Printf("Shortening link %q\n", link)
	w.WriteHeader(http.StatusCreated)
	linkid := router.Service.Create(link)
	_, err = io.WriteString(w, fmt.Sprintf("http://localhost:8080/%v", linkid))
	if err != nil {
		panic(err)
	}
}

func (router *URLRouter) handleGet(w http.ResponseWriter, r *http.Request) {
	linkid := chi.URLParam(r, "id")
	fmt.Printf("Getting link %q\n", linkid)

	if len(linkid) == 0 {
		http.Error(w, "Link id can't be empty", http.StatusBadRequest)
		return
	}

	link, ok := router.Service.Get(linkid)
	if !ok {
		http.Error(w, "Unknown link", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, link, http.StatusTemporaryRedirect)
}
