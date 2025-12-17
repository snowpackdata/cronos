package main

import (
	"html/template"
	"log"
	"net/http"
)

// CronosLandingHandler serves the Cronos platform landing page
func (a *App) CronosLandingHandler(w http.ResponseWriter, req *http.Request) {
	cronosTemplate, err := template.ParseFiles("./templates/cronos.html")
	if err != nil {
		log.Printf("Error parsing cronos template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = cronosTemplate.Execute(w, a.GitHash)
	if err != nil {
		log.Printf("Error executing cronos template: %v", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

// PasswordResetLandingHandler serves the password reset page
func (a *App) PasswordResetLandingHandler(w http.ResponseWriter, req *http.Request) {
	resetTemplate, err := template.ParseFiles("./templates/password_reset.html")
	if err != nil {
		log.Printf("Error parsing password reset template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = resetTemplate.Execute(w, a.GitHash)
	if err != nil {
		log.Printf("Error executing password reset template: %v", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

// NotFoundHandler serves the 404 error page
func (a *App) NotFoundHandler(w http.ResponseWriter, req *http.Request) {
	err404Template, err := template.ParseFiles("./templates/404.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	err = err404Template.Execute(w, a.GitHash)
	if err != nil {
		log.Printf("Error executing 404 template: %v", err)
	}
}

// BadRequestHandler serves the 400 error page
func (a *App) BadRequestHandler(w http.ResponseWriter, req *http.Request) {
	err400Template, err := template.ParseFiles("./templates/400.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	err = err400Template.Execute(w, a.GitHash)
	if err != nil {
		log.Printf("Error executing 400 template: %v", err)
	}
}
