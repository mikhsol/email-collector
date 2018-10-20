package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
)

type App struct {
	Router *mux.Router
	DB *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	var err error

	a.DB, err = sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
}

func (a *App) Run(address string) { }
