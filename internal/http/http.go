//This file is part of imdb2mkvtags ©2021-2026 Jörg Walter

package http

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/lcconv"
	"io"
	"net/http"
	"regexp"
)

var regexpLang = regexp.MustCompile("^[a-z]{2}(-[A-Z]{2})?$")
var internalClient = new(http.Client) //Default client for this library.

// Makes an HTTP request and writes the body to dest. If client is nil, the library's default client will be used.
func Body(client *http.Client, req *http.Request, dest io.Writer) error {
	if client == nil {
		client = internalClient
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

// Returns a new http request with default header fields
func NewBareReq(userAgent, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Charset", "utf-8")
	return req, nil
}

// Sets the passed language objects as the request's "Accept-Language" header field.
// Rejects formally malformed language strings to prevent header field injection.
func SetReqAccLang(req *http.Request, lang ...*lcconv.LngCntry) error {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for i, v := range lang {
		if v == nil {
			panic("argument \"lang\" cannot be nil")
		}
		if err := chkLang(v.HttpHeader()); err != nil {
			return err
		}
		if i == 0 {
			req.Header.Set("Accept-Language", v.HttpHeader())
		} else {
			req.Header.Add("Accept-Language", v.HttpHeader())
		}
	}
	return nil
}

// Makes an HTTP request to URL url, writes the answer's body to dest. If client is nil the library's default client will be used.
// If lang is not nil, the parameter will be used to set the request's Accept-Language parameter.
func GetBody(client *http.Client, userAgent, url string, dest io.Writer, lang ...*lcconv.LngCntry) error {
	req, err := NewBareReq(userAgent, "GET", url, nil)
	if err != nil {
		return err
	}
	if lang != nil {
		if err := SetReqAccLang(req, lang...); err != nil {
			return err
		}
	}
	return Body(client, req, dest)
}

func chkLang(s string) error {
	if !regexpLang.MatchString(s) {
		return errors.New("Malformed language string")
	}
	return nil
}
