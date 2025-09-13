package lfm

import "encoding/xml"

type Envelope struct {
	XMLName xml.Name `xml:"lfm"`
	Status  string   `xml:"status,attr"`
	Inner   []byte   `xml:",innerxml"`
}

type ApiError struct {
	Code    int    `xml:"code,attr"`
	Message string `xml:",chardata"`
}

type LastFMError struct {
	Code    int
	Message string
	Where   string
	Caller  string
}
