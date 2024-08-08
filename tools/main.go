package main

import (
	"fmt"
	"runtime"
)

func main() {
	os := runtime.GOOS
	arch := runtime.GOARCH

	fmt.Printf("%s_%s\n", os, arch)
}
