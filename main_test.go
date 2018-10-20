package main_test

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/mikhsol/email-collector"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)
import "testing"

var a main.App
var secret string

func TestMain(m *testing.M) {
	secret = os.Getenv("SECRET")
	a = main.App{}
	a.Initialize(
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

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestCreateCustomerSuccess(t *testing.T) {
	clearTable()
	name := "test user"
	email := "test@email.com"
	h := sha1.New()
	h.Write([]byte(strings.Join([]string{name, email, secret}, ":")))
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
	payload := []byte(fmt.Sprintf(`{"name":"%s","email":"%s","p":"%s"}`, name, email, sha))

	req, _ := http.NewRequest("POST", "/customer", bytes.NewBuffer(payload))

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != name {
		t.Errorf("Expected customer name to be 'test user'. Got '%v'", m["name"])
	}

	if m["email"] != email {
		t.Errorf("Expected customer email to be 'test@email.com'. Got '%v'", m["email"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected customer ID to be '1'. Got '%v'", m["id"])
	}
}

func TestCreateCustomerWrongHash(t *testing.T) {
	clearTable()
	payload := []byte(fmt.Sprintf(`{"name":"%s","email":"%s","p":"%s"}`,
		"test user", "test@email.com", "a"))

	req, _ := http.NewRequest("POST", "/customer", bytes.NewBuffer(payload))

	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if len(m) != 0 {
		t.Errorf("Expected customer no payload. Got '%v'", m)
	}
}

