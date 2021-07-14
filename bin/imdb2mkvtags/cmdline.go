//This file is part of imdb2mkvtags ©2021 Jörg Walter

package main

import (
	"flag"
)

type flags struct {
	out  *string  //output file
	lang *string  //language passed to the imdb website
	tail []string //non-processes args
}

func cmdFlags() *flags {
	f := new(flags)
	f.out = flag.String("o", "", "sets the output file")
	f.lang = flag.String("lang", "", "sets the preferred language for http requests")
	flag.Parse()
	f.tail = flag.Args()
	return f
}
