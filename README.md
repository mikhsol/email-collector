# Email Collector

Simple backend application to verify and collect customers emails.

The app working in two phases:

* `notify/` -- handle `POST` request with `email` and `name` fields.
Calculate hash from email and secret, create verification link, send
email to recipient with verification link using AWS SES.

* `customer/` -- handle `POST` request with  `email`, `name` and `p` fields. Calculate
hash from email and secret, compare with  paylode. On success store data in database.

Use sqlite to store data.

### Requirements

Golang version 1.10

### Customer model

Customer have next fields:
* ID
* Name
* Email

### Clone

`git clone git@github.com:mikhsol/email-collector.git`

### Build

* For dev purposes `go build`
* For production `go build -ldflags "-s -w"`

### Run

`./email-collector`

### Test
`go test`