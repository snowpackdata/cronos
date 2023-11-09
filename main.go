package main

import (
	"github.com/pkg/errors"
	"log"
)

func main() {
	a := &App{}
	a.Initialize()

	// Only want to seed and migrate on the initial build
	a.migrate()
	a.SeedDatabase()
	err := a.RegisterStaff("nathanaeljrobinson@gmail.com", 1)
	if err != nil && errors.Is(err, ErrUserAlreadyExists) {
		log.Print(ErrUserAlreadyExists.Error())
	}
}
