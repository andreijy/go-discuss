package web

import (
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
	h.Get("/", h.Home())
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", h.ThreadsList())
		r.Get("/new", h.ThreadsCreate())
		r.Post("/", h.ThreadsStore())
		r.Get("/{id}", h.ThreadsShow())
		r.Post("/{id}/delete", h.ThreadsDelete())
		r.Get("/{id}/new", h.PostsCreate())
		r.Post("/{id}", h.PostsStore())
		r.Get("/{threadID}/{postID}", h.PostsShow())
		r.Post("/{threadID}/{postID}", h.CommentsStore())
	})
	h.Get("/comments/{id}/vote", h.CommentsVote())

	return h
}

type Handler struct {
	*chi.Mux

	store godiscuss.Store
}

func (h *Handler) Home() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func (h *Handler) ThreadsList() http.HandlerFunc {
	type data struct {
		Threads []godiscuss.Thread
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/threads.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tt, err := h.store.Threads()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data{Threads: tt})
	}
}

func (h *Handler) ThreadsCreate() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/thread_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func (h *Handler) ThreadsShow() http.HandlerFunc {
	type data struct {
		Thread godiscuss.Thread
		Posts  []godiscuss.Post
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/thread.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		uuid, err := uuid.Parse(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		t, err := h.store.Thread(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pp, err := h.store.PostsByThread(t.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data{Thread: t, Posts: pp})
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

func (h *Handler) PostsCreate() http.HandlerFunc {
	type data struct {
		Thread godiscuss.Thread
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/post_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		uuid, err := uuid.Parse(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		t, err := h.store.Thread(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data{Thread: t})
	}
}

func (h *Handler) PostsShow() http.HandlerFunc {
	type data struct {
		Post     godiscuss.Post
		Comments []godiscuss.Comment
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/post.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "postID")
		postID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p, err := h.store.Post(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cc, err := h.store.CommentsByPost(p.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data{Post: p, Comments: cc})
	}
}

func (h *Handler) PostsStore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		content := r.FormValue("content")

		idStr := chi.URLParam(r, "id")
		threadID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(threadID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p := &godiscuss.Post{
			ID:       uuid.New(),
			ThreadID: t.ID,
			Title:    title,
			Content:  content,
		}

		err = h.store.CreatePost(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/threads/"+t.ID.String()+"/"+p.ID.String(), http.StatusFound)
	}
}

func (h *Handler) CommentsStore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content := r.FormValue("content")
		idStr := chi.URLParam(r, "postID")

		postID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c := &godiscuss.Comment{
			ID:      uuid.New(),
			PostID:  postID,
			Content: content,
		}

		err = h.store.CreateComment(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
}

func (h *Handler) CommentsVote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c, err := h.store.Comment(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		dir := r.URL.Query().Get("dir")
		if dir == "up" {
			c.Votes++
		} else if dir == "down" {
			c.Votes--
		}

		err = h.store.UpdateComment(&c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
}
