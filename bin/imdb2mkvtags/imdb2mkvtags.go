//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	"github.com/jwdev42/imdb2mkvtags/internal/controller"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"os"
)

func write(file *os.File, data *tags.Movie) {
	if file != os.Stdout {
		defer file.Close()
	}
	if err := tags.WriteTags(file, data.WriteTag); err != nil {
		global.Log.Die(fmt.Errorf("Error writing output: %s", err))
	}
}

func main() {

	flags, err := cmdline.Parse()
	if err != nil {
		global.Log.Die(fmt.Errorf("Error parsing command line: %s", err))
	}
	if *flags.LegalInfo {
		printLegalInfo()
	}

	if len(flags.Tail) < 1 {
		global.Log.Die("No URL specified in input")
	}
	c, err := controller.Pick(flags.Tail[0])
	if err != nil {
		global.Log.Die(err)
	}
	if err := c.SetOptions(flags); err != nil {
		global.Log.Die(fmt.Errorf("Could not set scraper options: %s", err))
	}
	movie, err := c.Scrape()
	if err != nil {
		global.Log.Die(fmt.Errorf("Scraping error: %s", err))
	}
	var file *os.File
	if *flags.Out != "" {
		var err error
		file, err = os.Create(*flags.Out)
		if err != nil {
			global.Log.Die(err)
		}
	} else {
		file = os.Stdout
	}
	write(file, movie)
}
