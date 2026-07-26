package handlers

import (
	"fmt"
	"runtime"
)

func HandleVersion(version string) {
	fmt.Printf("cwclock-cli/%v %v %v\n", version, runtime.GOOS, runtime.GOARCH)
}
