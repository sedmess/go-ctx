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

//goland:noinspection GoUnusedExportedFunction
func String() string {
	if //goland:noinspection GoBoolExpressions
	Version == "" {
		return Name
	} else {
		return Name + " " + Version
	}
}

//goland:noinspection GoUnusedExportedFunction
func FullString() string {
	str := String()
	if //goland:noinspection GoBoolExpressions
	BuildInfo != "" {
		str += " (" + BuildInfo + ")"
	}
	return str
}
