package main

import (
	"log"

	"github.com/tamsanh/go_lock/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
