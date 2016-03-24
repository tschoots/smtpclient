package common

import (

)

type Attachment struct {
	FilePath string
	FileName string
}

type Mail struct {
	Name         string
	Surname      string
	Company      string
	ToEMail      string
	Subject      string
	FromEmail    string
	Body         string
	Attachments  []Attachment
}
