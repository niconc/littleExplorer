package main

import (
	"fmt"
	"littleExplorer/server"
	"log"
)

func init() {
	fmt.Println()
	log.Println("__INIT__: main() NASA API explorer initiallization. init() is running.")
}

func main() {
	setupApodAPI()  // API Information.
	server.Server() // run the server.
}

func setupApodAPI() {
	var apodID []string // the identification values of apod.
	apodID = []string{"APOD", "Astronomy Picture of the Day!"}
	log.Println("Low Nasa Orbit: Bring NASA to people!")
	log.Printf("%s %s\n\n", apodID[0], apodID[1])
}
