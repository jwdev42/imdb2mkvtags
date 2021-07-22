//This file is part of imdb2mkvtags ©2021 Jörg Walter

package cmdline

import (
	"flag"
)

//Structure that holds the parsed command line flags
type Flags struct {
	Out  *string  //output file
	Lang *string  //language passed to the website
	Imdb *string  //options for the imdb scraper
	Tail []string //non-processed args
}

func Parse() *Flags {
	f := new(Flags)
	f.Out = flag.String("o", "", "Sets the output file.")
	f.Lang = flag.String("lang", "", "Sets the preferred language(s) for http requests. Multiple languages are separated by a colon.")
	f.Imdb = flag.String("imdb", "", "Options for the imdb scraper, separated by a colon.")
	flag.Parse()
	f.Tail = flag.Args()
	return f
}
