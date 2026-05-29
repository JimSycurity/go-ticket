package main

import (
	"os"

	"github.com/JimSycurity/go-ticket/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
