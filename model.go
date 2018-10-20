package main

import (
	"database/sql"
	"errors"
)

type customer struct {
	ID    int    "json:'id'"
	Name  string "json:'name'"
	Email string "json:'email'"
}

func (c *customer) getCustomer(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (c *customer) createCustomer(db *sql.DB) error {
	return errors.New("Not implemented")
}

func getCustomers(db *sql.DB, start, count int) ([]customer, error) {
	return nil, errors.New("Not implemented")
}
