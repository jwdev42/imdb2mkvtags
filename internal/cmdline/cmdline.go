//This file is part of imdb2mkvtags ©2021 Jörg Walter

package cmdline

import (
	"flag"
)

//Structure that holds the parsed command line flags
type Flags struct {
	Out  *string  //output file
	Lang *string  //language passed to the website
	Opts *string  //options for the scraper
	Tail []string //non-processed args
}

func Parse() *Flags {
	f := new(Flags)
	f.Out = flag.String("o", "", "Sets the output file.")
	f.Lang = flag.String("lang", "", "Sets the preferred language(s) for http requests. Multiple languages are separated by a colon.")
	f.Opts = flag.String("opts", "", "Scraper-specific options, separated by a colon.")
	flag.Parse()
	f.Tail = flag.Args()
	return f
}
