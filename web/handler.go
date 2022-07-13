package web

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	godiscuss "github.com/andreijy/go-discuss"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func NewHandler(store godiscuss.Store, sessions *scs.SessionManager) *Handler {
	h := &Handler{
		Mux:      chi.NewMux(),
		store:    store,
		sessions: sessions,
	}

	threadHandler := ThreadHandler{store: store, sessions: sessions}
	postHandler := PostHandler{store: store, sessions: sessions}
	commentHandler := CommentHandler{store: store, sessions: sessions}
	userHandler := UserHandler{store: store, sessions: sessions}

	csrfKey := []byte("01234567890123456789012345678901")

	h.Use(middleware.Logger)
	// TODO: for https only, replace with
	// TODO: h.Use(csrf.Protect(csrfKey))
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false)))
	h.Use(sessions.LoadAndSave)

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
	h.Get("/register", userHandler.Register())
	h.Post("/register", userHandler.RegisterSubmit())

	return h
}

type Handler struct {
	*chi.Mux

	store    godiscuss.Store
	sessions *scs.SessionManager
}

func (h *Handler) Home() http.HandlerFunc {
	type data struct {
		SessionData
		Posts []godiscuss.Post
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		pp, err := h.store.Posts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			Posts:       pp,
		})
	}
}
