package main_test

import (
	"github.com/mikhsol/email-collector"
	"log"
	"os"
)
import "testing"

var a main.App

func MainTest(m *testing.M) {

	os.Setenv("ENV", "TEST")

	a = main.App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func clearTable() {
	a.DB.Exec("DELETE FROM customers")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS customers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(128) NOT NULL,
  email varchar(128) NOT NULL)`


func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}
