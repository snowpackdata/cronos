package main

func main() {
	a := &App{}
	a.Initialize()

	// Only want to seed and migrate on the initial build
	// a.migrate()
	// a.SeedDatabase()

}
