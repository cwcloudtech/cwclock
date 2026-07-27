// Package main provides an entry point for cwclock application.
package main

import "cwclock/cmd"

// Version represents the application version.
var Version = "dev"

func main() {
	cmd.Execute(Version)
}
