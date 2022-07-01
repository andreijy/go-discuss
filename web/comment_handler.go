package web

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	godiscuss "github.com/andreijy/go-discuss"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CommentHandler struct {
	store    godiscuss.Store
	sessions *scs.SessionManager
}

func (h *CommentHandler) Store() http.HandlerFunc {
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

		h.sessions.Put(r.Context(), "flash", "Your comment has beed added.")

		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
}

func (h *CommentHandler) Vote() http.HandlerFunc {
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
