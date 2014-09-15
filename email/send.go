package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func main() {
	c, err := smtp.Dial("mail.it.northwestern.edu:25")
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Mail("siddhartha-basu@northwestern.edu"); err != nil {
		log.Fatalf("Error in mail command %s\n", err)
	}
	if err := c.Rcpt("basusiddhartha@gmail.com"); err != nil {
		log.Fatal(err)
	}

	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, "This is a test mail from golang")
	if err != nil {
		log.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = c.Quit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent the email")
}
