package appinfo

import (
	"os"
	"path/filepath"
)

var Name = ""
var Version = ""
var BuildInfo = ""

func init() {
	if Name == "" {
		Name = filepath.Base(os.Args[0])
	}
}

func String() string {
	if //goland:noinspection GoBoolExpressions
	Version == "" {
		return Name
	} else {
		return Name + " " + Version
	}
}
