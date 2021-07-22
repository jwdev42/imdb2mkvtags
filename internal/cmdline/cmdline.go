//This file is part of imdb2mkvtags ©2021 Jörg Walter

package cmdline

import (
	"flag"
)

type Flags struct {
	Out  *string  //output file
	Lang *string  //language passed to the imdb website
	Tail []string //non-processes args
}

func Parse() *Flags {
	f := new(Flags)
	f.Out = flag.String("o", "", "sets the output file")
	f.Lang = flag.String("lang", "", "sets the preferred language for http requests")
	flag.Parse()
	f.Tail = flag.Args()
	return f
}
