package main

func main() {
	a := &App{}
	a.Initialize()
	a.migrate()
	a.SeedDatabase()
}
