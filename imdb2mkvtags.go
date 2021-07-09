//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"bytes"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/modhttp"
	"github.com/jwdev42/imdb2mkvtags/modimdb"
	"github.com/jwdev42/imdb2mkvtags/modtag"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	url, err := modimdb.Id2Url(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid Input: %s\n", err)
		os.Exit(1)
	}
	buf := new(bytes.Buffer)
	if err := modhttp.Body(url, buf); err != nil {
		fmt.Fprintf(os.Stderr, "Error receiving webpage: %s\n", err)
		os.Exit(1)
	}
	sMovie, err := modimdb.ExtractMovieSchema(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while extracting movie schema: %s\n", err)
		os.Exit(1)
	}
	movie := sMovie.Convert()
	if err := modtag.WriteTags(os.Stdout, movie.WriteTag); err != nil {
		fmt.Fprintf(os.Stderr, "Error on xml output: %s\n", err)
		os.Exit(1)
	}
}
