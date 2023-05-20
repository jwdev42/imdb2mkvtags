//This file is part of imdb2mkvtags ©2021 Jörg Walter

package cmdline

import (
	"flag"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/lcconv"
	"github.com/jwdev42/logger"
	"strings"
)

// Structure that holds the parsed command line flags
type Flags struct {
	LegalInfo *bool //Print legal info?
	Loglevel  logger.LevelFlag
	Out       *string //output file
	rawLang   *string //language-country combination(s)
	Lang      []*lcconv.LngCntry
	Opts      *string  //options for the scraper
	Tail      []string //non-processed args
}

func Parse() (*Flags, error) {
	f := &Flags{Loglevel: logger.LevelFlag(global.DefaultLoglevel)}
	f.LegalInfo = flag.Bool("print-legal-info", false, "Print legal information and exit.")
	f.Out = flag.String("o", "", "Sets the output file.")
	f.rawLang = flag.String("lang", "en-US", "Sets the preferred language(s) for http requests. Multiple languages are separated by a colon.")
	f.Opts = flag.String("opts", "", "Scraper-specific options, separated by a colon.")
	flag.Var(&f.Loglevel, "loglevel", "set the logging verbosity.")
	flag.Parse()
	if err := f.parseLang(); err != nil {
		return nil, err
	}
	f.Tail = flag.Args()
	global.Log.SetLevel(int(f.Loglevel))
	return f, nil
}

func (r *Flags) parseLang() error {
	rawLangs := strings.Split(*r.rawLang, ":")
	langs := make([]*lcconv.LngCntry, len(rawLangs))
	for i, rawLang := range rawLangs {
		lang, err := lcconv.NewLngCntry(rawLang)
		if err != nil {
			return err
		}
		langs[i] = lang
	}
	r.Lang = langs
	return nil
}
