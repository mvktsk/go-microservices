package main

// Package app provides the main functionality for the UserService application.
// It handles HTTP requests, interacts with the database, and performs various operations related to user management.
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type App struct {	
	DB	*sqlx.DB
	Router	*mux.Router
}

func (a *App) Initialize(db *sqlx.DB) {
	a.DB = db
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes(){
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/user", a.createUser).Methods("POST")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

func (a *App) Run(addr string) {
	n := negroni.Classic()
	n.UseHandler(a.Router)
	log.Fatal(http.ListenAndServe(addr, n))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func responseWithError(w http.ResponseWriter, code int, message string){
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	user := User{ID: id}
	err = user.get(a.DB)
	if err != nil {
		switch err {
			case sql.ErrNoRows:
				responseWithError(w, http.StatusNotFound, "User not found ID: " + strconv.Itoa(id))
			default:
				responseWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request){
	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := user.create(a.DB); err != nil {
		fmt.Println(err.Error())
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	user.ID = id
	if err := user.update(a.DB); err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	user := User{ID: id}
	if err := user.delete(a.DB); err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}