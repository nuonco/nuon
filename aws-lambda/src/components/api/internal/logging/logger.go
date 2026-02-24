package logging

import (
	"log"
	"os"
)

var DefaultLogger = log.New(os.Stdout, "", log.Llongfile)
