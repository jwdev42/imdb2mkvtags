//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"bytes"
	"errors"
	"fmt"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"os"
	"strings"
)

func die(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func write(file *os.File, data *tags.Movie) {
	if file != os.Stdout {
		defer file.Close()
	}
	if err := tags.WriteTags(file, data.WriteTag); err != nil {
		die(fmt.Errorf("Error writing output: %s", err))
	}
}

func main() {

	flags := cmdFlags()

	if len(flags.tail) < 1 {
		die(errors.New("No IMDB title id found"))
	}
	url, err := imdb.Id2Url(flags.tail[0])
	if err != nil {
		die(fmt.Errorf("Invalid Input: %s", err))
	}
	req, err := ihttp.NewBareReq("GET", url, nil)
	if err != nil {
		die(fmt.Errorf("Could not create HTTP request: %s", err))
	}
	if len(*flags.lang) > 0 {
		if err := ihttp.SetReqAccLang(req, strings.Split(*flags.lang, ",")...); err != nil {
			die(fmt.Errorf("Could not set header for \"Accept-Language\": %s", err))
		}
	}
	buf := new(bytes.Buffer)
	if err := ihttp.Body(nil, req, buf); err != nil {
		die(fmt.Errorf("Error receiving webpage: %s", err))
	}
	sMovie, err := imdb.ExtractMovieSchema(buf)
	if err != nil {
		die(fmt.Errorf("Error while extracting movie schema: %s", err))
	}
	var file *os.File
	if *flags.out != "" {
		var err error
		file, err = os.Create(*flags.out)
		if err != nil {
			die(err)
		}
	} else {
		file = os.Stdout
	}
	write(file, sMovie.Convert())
}
