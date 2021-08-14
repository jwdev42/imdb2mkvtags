//This file is part of imdb2mkvtags ©2021 Jörg Walter

package global

import (
	"github.com/jwdev42/logger"
	"os"
)

const DefaultLoglevel = logger.LevelNotice
const DelimControllerArgs = ":" //delimiter to separate controller-specific arguments.
const DelimControllerKV = "="   //delimiter to separate arguments and values in controller-specific arguments.

var Log *logger.Logger = logger.New(os.Stderr, DefaultLoglevel, " - ")
