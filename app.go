package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

type App struct {
	Router *mux.Router
	DB *sql.DB
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS customers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(128) NOT NULL,
  email varchar(128) NOT NULL)`

func (a *App) Initialize(dbname string) {
	var err error

	a.DB, err = sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/customer", a.createCustomer).Methods("POST")
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) createCustomer(w http.ResponseWriter, r *http.Request) {
	var c customer
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		log.Printf("CREATE CUSTOMER: %e", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if c.toSha1(os.Getenv("SECRET")) != c.P {
		log.Printf("CREATE CUSTOMER: Wrong hash %s for %s", c.P, c.Email)
		respondWithJSON(w, http.StatusOK, nil)
		return
	}

	if err := c.createCustomer(a.DB); err != nil {
		log.Printf("CREATE CUSTOMER: %e", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("CREATE CUSTOMER: customer with id %d and email %s was created",
		c.ID, c.Email)
	respondWithJSON(w, http.StatusCreated, c)
}
