//This file is part of imdb2mkvtags ©2021 Jörg Walter

package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const UserAgent = "imdb2mkvtags/1.0"

var regexpLang = regexp.MustCompile("^[a-z]{2}(-[A-Z]{2})?$")

func Body(client *http.Client, req *http.Request, dest io.Writer) error {
	if client == nil {
		client = new(http.Client)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(dest, resp.Body); err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("HTTP response: %s", resp.Status))
	}

	return nil
}

//Returns a new http request with default header fields
func NewBareReq(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept-Charset", "utf-8")
	return req, nil
}

//Sets the passed language strings as the request's "Accept-Language" header field.
//Rejects formally malformed language strings to prevent header field injection.
func SetReqAccLang(req *http.Request, lang ...string) error {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for _, v := range lang {
		if err := chkLang(v); err != nil {
			return err
		}
	}
	req.Header.Set("Accept-Language", strings.Join(lang, ","))
	return nil
}

func chkLang(s string) error {
	if !regexpLang.MatchString(s) {
		return errors.New("Malformed language string")
	}
	return nil
}
