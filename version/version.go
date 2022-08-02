package version

import (
	"fmt"
	"os"
)

// Version overlord version consts
var Version = "1.0.0"

// ShowVersion print version if -version flag is seted and return true
func ShowVersion() bool {
	fmt.Fprintln(os.Stdout, Version)
	return true
}

// Bytes return version bytes
func Bytes() []byte {
	return []byte(Version)
}

// Str is the formatted version string
func Str() string {
	return Version
}
