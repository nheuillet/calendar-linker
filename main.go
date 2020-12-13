package main

import (
	"log"

	"github.com/nheuillet/calendar-linker/agenda"
	"github.com/nheuillet/calendar-linker/intra"
	"github.com/nheuillet/calendar-linker/parser"
)

func handleErrors(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	projects := &[]intra.Activity{}
	registeredEvents := &[]intra.Event{}
	config, err := parser.GetConfigInfos()
	handleErrors(err)
	err = intra.GetRegisteredEvents(config, registeredEvents)
	handleErrors(err)

	if config.ProjectEvent {
		// if ProjectEvent is set to True then it will fetch all the modules
		// in order to retrieve the Project information, then it will add it to
		// the agenda.
		// WARNING: fetching every project is very long due to the very bad REST api of the intra..
		// UPDATE: it is  no longer long. Going for a freaking huge amount of goroutine does the trick. Intra Api is still very bad.
		err = intra.GetProjects(config, projects)
		handleErrors(err)
	}
	googleClient := agenda.GetGoogleClient()

	agenda.CreateEvents(googleClient, config,
		registeredEvents, projects)
}
