package smtpclient

import (
	"log"
	"net"
	"bytes"
	"net/smtp"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"encoding/base64"
)

type SmtpClient struct {
	Servername string
	SmtpUser   string
	SmtpPasswd string
	Mail       Mail
}

func (s *SmtpClient) Send() {

	marker := "ACUSTOMANDUNIQUEBOUNDARY"

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = s.Mail.FromEmail
	headers["To"] = s.Mail.ToEMail
	headers["Subject"] = s.Mail.Subject
	headers["Content-Type"] = "multipart/mixed"

	//part 1 will be the mail headers
	part1 := fmt.Sprintf("From: Black Duck Software OSRP Report %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n--%s", s.Mail.FromEmail, s.Mail.ToEMail, s.Mail.Subject, marker, marker)
	fmt.Printf("%s\n", part1)

	//part 2 will be the body of the email (text or HTML)
	part2 := fmt.Sprintf("\r\n\r\n--%s\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\ncharset=\"UTF-8\"\r\n\r\n%s\r\n--%s", marker, s.Mail.Body, marker)
	fmt.Printf("\n\n%s\n\n", part2)

	message := part1 + part2

	// check if there are attachments
	if s.Mail.Attachments != nil {
	for _, attachment := range *s.Mail.Attachments {

		// read and encode attachment
		content, _ := ioutil.ReadFile(attachment.FilePath)
		encoded := base64.StdEncoding.EncodeToString(content)

		//split the encoded file in lines (doesn matter, but low enough not to hit max limit)
		lineMaxLength := 500
		nbrLines := len(encoded) / lineMaxLength

		// create a buffer
		var buf bytes.Buffer
		for i := 0; i < nbrLines; i++ {
			buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
		} //for
		// append the last line to buffer
		buf.WriteString(encoded[nbrLines*lineMaxLength:])

		//part 3 will be the attachment
		part3 := fmt.Sprintf("\r\nContent-Type: application/csv; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n\r\n--%s", attachment.FilePath, attachment.FileName, buf.String(), marker)
		message = message + part3
	}
	}

	message = message + "--"
	fmt.Printf("mail:\n%s\n\n", message)

	

	// Connect to the SMTP Server
	host, _, _ := net.SplitHostPort(s.Servername)

	auth := smtp.PlainAuth("", s.SmtpUser, s.SmtpPasswd, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", s.Servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To & From
	if err = c.Mail(s.Mail.FromEmail); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(s.Mail.ToEMail); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	c.Quit()
}
