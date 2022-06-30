package web

import (
	"html/template"
	"net/http"

	godiscuss "github.com/andreijy/go-discuss"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func NewHandler(store godiscuss.Store) *Handler {
	h := &Handler{
		Mux:   chi.NewMux(),
		store: store,
	}
	threadHandler := ThreadHandler{store: store}
	postHandler := PostHandler{store: store}
	commentHandler := CommentHandler{store: store}

	csrfKey := []byte("01234567890123456789012345678901")

	h.Use(middleware.Logger)

	// TODO: for https only, replace with
	// TODO: h.Use(csrf.Protect(csrfKey))
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false)))
	h.Get("/", h.Home())
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", threadHandler.List())
		r.Get("/new", threadHandler.Create())
		r.Post("/", threadHandler.Store())
		r.Get("/{id}", threadHandler.Show())
		r.Post("/{id}/delete", threadHandler.Delete())
		r.Get("/{id}/new", postHandler.Create())
		r.Post("/{id}", postHandler.Store())
		r.Get("/{threadID}/{postID}", postHandler.Show())
		r.Get("/{threadID}/{postID}/vote", postHandler.Vote())
		r.Post("/{threadID}/{postID}", commentHandler.Store())
	})
	h.Get("/comments/{id}/vote", commentHandler.Vote())

	return h
}

type Handler struct {
	*chi.Mux

	store godiscuss.Store
}

func (h *Handler) Home() http.HandlerFunc {
	type data struct {
		Posts []godiscuss.Post
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		pp, err := h.store.Posts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data{Posts: pp})
	}
}
