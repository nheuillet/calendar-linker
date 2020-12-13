package parser

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// Config structure that holds the data provided by the user in the configuration file
type Config struct {
	GoogleCalendarEvents   string `json:"google_calendar_events"`      // the calendar id of daily events. default calendar is "Primary"
	GoogleCalendarProjects string `json:"google_calendar_projects"`    // the calendar id of the project events. You can use the same id as the upper field. default calendar is "Primary"
	EpitechAuth            string `json:"epitech_auth"`                // the autologin token. Starts with "auth-"
	Location               string `json:"epitech_location_code"`       // Location code on the intra. Eg: FR/TLS for Toulouse (<3)
	ProjectEvent           bool   `json:"create_project_event"`        // If you want to create the project events on your calendar
	ProjectParticipant     bool   `json:"add_participants_to_project"` // Turning it to true adds participants to the project. Caution: leads to N * More call to the API, N being the number of projects.
	Semesters              []int  `json:"epitech_semesters"`           // Semesters you want to scan if ProjectEvent set to true.
	Timezone               string `json:"timezone"`                    // The timezone of the epitech you are enrolled in
	ProjectColor           string `json:"project_color"`               // see https://lukeboyle.com/blog-posts/2016/04/google-calendar-api---color-id
	EventColor             string `json:"event_color"`                 // see https://lukeboyle.com/blog-posts/2016/04/google-calendar-api---color-id
	Reminders              []int  `json:"reminder_time"`               // Array Number of minutes in order to get a notification. Max is 40320 per google api recommendation(4 weeks in minutes)
	LocationRegex          string `json:"location_regex"`              // The regex used to extract the name of the room. Using a regex in order to make sure it is customizable for every Epitech. Epitech Toulouse example is FR/TLS/Marquette/ROOMNAME
}

// GetConfigInfos will create a config instance containing all the data from the config file
func GetConfigInfos() (*Config, error) {
	file, err := ioutil.ReadFile("config.json")
	var conf Config

	if err != nil {
		return nil, errors.New(("could not open file"))
	}
	err = json.Unmarshal([]byte(file), &conf)
	if err != nil {
		return nil, errors.New("could not unmarshall json, make sure the config file is in a correct format")
	}
	return &conf, nil
}
