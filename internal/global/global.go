package global

import (
	"github.com/jwdev42/logger"
	"os"
)

const DefaultLoglevel = logger.LevelError

var Log *logger.Logger = logger.New(os.Stderr, DefaultLoglevel, ": ")
