package main

import (
	"database/sql"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strings"
)

type App struct {
	Router         *mux.Router
	DB             *sql.DB
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
	a.Router.HandleFunc("/notify", a.notifyCustomer).Methods("POST")
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

func (a *App) notifyCustomer(w http.ResponseWriter, r *http.Request) {
	var c customer
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		log.Printf("CREATE CUSTOMER: %e", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	verificationLink := strings.Join([]string{os.Getenv("BASE_URL"),
		"?name=", c.Name, "&email=", c.Email, "&p=",
		c.toSha1(os.Getenv("SECRET"))}, "")

	sender := os.Getenv("SENDER")
	recipient := c.Email
	subject := os.Getenv("SUBJECT")

	// The HTML body for the email.
	htmlBody :=  strings.Join([]string{
		"<h1>Dear ", c.Name, "</h1><p>Please, press next url to confirm your email ",
		"<a href='", verificationLink, "'>Awesome verification link</a>"}, "")

	//The email body for recipients with non-HTML email clients.
	textBody := strings.Join([]string{
		"Dear ", c.Name, "\nPlease, press next url to confirm your email ",
		verificationLink, " ."}, "")

	charSet := "UTF-8"

	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region:aws.String(os.Getenv("AWS_REGION"))},
	)

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		respondWithJSON(w, http.StatusInternalServerError, nil)
	}

	log.Println("Email Sent to address: " + recipient)
	log.Println(result)

	respondWithJSON(w, http.StatusOK, nil)
}
