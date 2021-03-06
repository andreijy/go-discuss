package web

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	godiscuss "github.com/andreijy/go-discuss"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

type PostHandler struct {
	store    godiscuss.Store
	sessions *scs.SessionManager
}

// To handle r.Get("/{id}/new", postHandler.Create())
func (h *PostHandler) Create() http.HandlerFunc {
	type data struct {
		SessionData
		Thread godiscuss.Thread
		CSRF   template.HTML
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

		tmpl.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			Thread:      t,
			CSRF:        csrf.TemplateField(r),
		})
	}
}

// To handle r.Get("/{threadID}/{postID}", postHandler.Show())
func (h *PostHandler) Show() http.HandlerFunc {
	type data struct {
		SessionData
		Post     godiscuss.Post
		Comments []godiscuss.Comment
		CSRF     template.HTML
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

		tmpl.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			Post:        p,
			Comments:    cc,
			CSRF:        csrf.TemplateField(r),
		})
	}
}

// To handle r.Post("/{id}", postHandler.Store())
func (h *PostHandler) Store() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := CreatePostForm{
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		}

		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(w, r, r.Referer(), http.StatusFound)
			return
		}

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
			Title:    form.Title,
			Content:  form.Content,
		}

		err = h.store.CreatePost(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your new post has been created.")

		http.Redirect(w, r, "/threads/"+t.ID.String()+"/"+p.ID.String(), http.StatusFound)
	}
}

// To handle r.Get("/{threadID}/{postID}/vote", postHandler.Vote())
func (h *PostHandler) Vote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "postID")

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p, err := h.store.Post(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		dir := r.URL.Query().Get("dir")
		if dir == "up" {
			p.Votes++
		} else if dir == "down" {
			p.Votes--
		}

		err = h.store.UpdatePost(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
}
