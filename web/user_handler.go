package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	godiscuss "github.com/andreijy/go-discuss"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	store    godiscuss.Store
	sessions *scs.SessionManager
}

// To handle h.Get("/register", userHandler.Register())
func (h *UserHandler) Register() http.HandlerFunc {
	type data struct {
		SessionData
		CSRF template.HTML
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/user_register.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("here0")

		tmpl.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			CSRF:        csrf.TemplateField(r)})
	}
}

// To handle h.Post("/register", userHandler.RegisterSubmit())
func (h *UserHandler) RegisterSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("here0")

		form := RegisterForm{
			Username:      r.FormValue("username"),
			Password:      r.FormValue("password"),
			UsernameTaken: false,
		}
		fmt.Println("here1")
		_, err := h.store.UserByUsername(form.Username)
		if err == nil {
			form.UsernameTaken = true
		}

		fmt.Println("here2")

		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(w, r, r.Referer(), http.StatusFound)
			return
		}
		fmt.Println("here3")

		// hash the password with bcrypt
		password, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("here4")

		err = h.store.CreateUser(&godiscuss.User{
			ID:       uuid.New(),
			Username: form.Username,
			Password: string(password),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your registration was successful. Please log in.")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
