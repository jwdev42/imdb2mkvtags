//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"bytes"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"os"
)

func write(file *os.File, data *tags.Movie) {
	if file != os.Stdout {
		defer file.Close()
	}
	if err := tags.WriteTags(file, data.WriteTag); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %s\n", err)
		os.Exit(1)
	}
}

func main() {

	flags := cmdFlags()

	if len(flags.tail) < 1 {
		fmt.Fprintf(os.Stderr, "No IMDB title id found\n")
		os.Exit(0)
	}
	url, err := imdb.Id2Url(flags.tail[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid Input: %s\n", err)
		os.Exit(1)
	}
	buf := new(bytes.Buffer)
	if err := http.Body(url, buf); err != nil {
		fmt.Fprintf(os.Stderr, "Error receiving webpage: %s\n", err)
		os.Exit(1)
	}
	sMovie, err := imdb.ExtractMovieSchema(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while extracting movie schema: %s\n", err)
		os.Exit(1)
	}
	var file *os.File
	if *flags.out != "" {
		var err error
		file, err = os.Create(*flags.out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else {
		file = os.Stdout
	}
	write(file, sMovie.Convert())
}
