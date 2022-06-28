package web

import (
	"fmt"
	"html/template"
	"net/http"

	godiscuss "github.com/andreijy/go-discuss"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func NewHandler(store godiscuss.Store) *Handler {
	h := &Handler{
		Mux:   chi.NewMux(),
		store: store,
	}

	h.Use(middleware.Logger)
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", h.ThreadsList())
		r.Get("/new", h.ThreadsCreate())
		r.Post("/", h.ThreadsStore())
		r.Post("/{id}/delete", h.ThreadsDelete())
	})

	h.Get("/html", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("layout.html").ParseGlob("templates/includes/*.html"))
		t = template.Must(t.ParseFiles("templates/layout.html", "templates/childtemplate.html"))

		type params struct {
			Title   string
			Text    string
			Lines   []string
			Number1 int
			Number2 int
		}

		t.Execute(w, params{
			Title: "Golang discussion board",
			Text:  "Welcome to our discussion board",
			Lines: []string{
				"Lines 1",
				"Lines 2",
				"Lines 3",
			},
			Number1: 33,
			Number2: 33,
		})
	})

	return h
}

type Handler struct {
	*chi.Mux

	store godiscuss.Store
}

const threadsListHTML = `
<h1>Threads</h1>
<dl>
{{range .}}
	<dt><strong>{{.Title}}</strong></dt>
	<dd>{{.Description}}</dd>
	<dd>
	<form action="/threads/{{.ID}}/delete" method="POST">
		<button type="submit">Delete</button>
	</form>
	</dd>
{{end}}
</dl>
<a href="/threads/new">Create thread</a>
`

func (h *Handler) ThreadsList() http.HandlerFunc {
	tmpl := template.Must(template.New("").Parse(threadsListHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		tt, err := h.store.Threads()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, tt)
	}
}

const threadCreateHTML = `
<h1>New Thread</h1>
<form action="/threads" method="POST">
	<table>
		<tr>
			<td>Title</td>
			<td><input type="text" name="title"></td>
		</tr>
		<tr>
			<td>Description</td>
			<td><input type="text" name="description"></td>
		</tr>
	</table>
	<button type="submit">Create thread</button>
</form>
`

func (h *Handler) ThreadsCreate() http.HandlerFunc {
	tmpl := template.Must(template.New("").Parse(threadCreateHTML))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func (h *Handler) ThreadsStore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		description := r.FormValue("description")

		err := h.store.CreateThread(&godiscuss.Thread{
			ID:          uuid.New(),
			Title:       title,
			Description: description,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/threads", http.StatusFound)
	}
}

func (h *Handler) ThreadsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		fmt.Println("delete")
		uuid, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = h.store.DeleteThread(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/threads", http.StatusFound)
	}
}
