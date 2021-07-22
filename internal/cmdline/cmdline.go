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
	f.Out = flag.String("o", "", "sets the output file")
	f.Lang = flag.String("lang", "", "sets the preferred language for http requests")
	f.Imdb = flag.String("imdb", "", "options for the imdb scraper, separated by a colon")
	flag.Parse()
	f.Tail = flag.Args()
	return f
}
