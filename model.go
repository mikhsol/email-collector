package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"log"
	"strings"
)

type customer struct {
	ID    int64   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	P     string `json:"p"`
}

func (c *customer) createCustomer(db *sql.DB) error {
	stmt, err := db.Prepare("INSERT INTO customers (name, email) values(?,?)")
	if err != nil {
		log.Println(err)
		return err
	}

	res, err := stmt.Exec(c.Name, c.Email)
	if err != nil {
		log.Println(err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return err
	}
	c.ID = id

	return nil
}

func (c *customer) toSha1(secret string) string {
	h := sha1.New()
	h.Write([]byte(strings.Join([]string{c.Name, c.Email, secret}, ":")))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}