//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	"github.com/jwdev42/imdb2mkvtags/internal/controller"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"os"
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

	flags := cmdline.Parse()

	if len(flags.Tail) < 1 {
		die(errors.New("No URL found"))
	}
	c, err := controller.Pick(flags.Tail[0])
	if err != nil {
		die(err)
	}
	movie, err := c.Scrape()
	if err != nil {
		die(fmt.Errorf("Error while extracting movie schema: %s", err))
	}
	var file *os.File
	if *flags.Out != "" {
		var err error
		file, err = os.Create(*flags.Out)
		if err != nil {
			die(err)
		}
	} else {
		file = os.Stdout
	}
	write(file, movie)
}
