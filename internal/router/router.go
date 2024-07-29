package router

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/config"
)

type URLRouter struct {
	chi.Router
	Service Service
	Flags   config.Flags
}

type Service interface {
	Create(value string) (ID string)
	Get(ID string) (string, bool)
}

func NewURLRouter(s Service, c config.Flags) *URLRouter {
	r := URLRouter{chi.NewRouter(), s, c}

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

	w.WriteHeader(http.StatusCreated)
	linkid := router.Service.Create(link)
	shorturl, err := url.JoinPath(router.Flags.RedirectBaseAddr, linkid)
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, shorturl)
	if err != nil {
		panic(err)
	}
}

func (router *URLRouter) handleGet(w http.ResponseWriter, r *http.Request) {
	linkid := chi.URLParam(r, "id")

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
