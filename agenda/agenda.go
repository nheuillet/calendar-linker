package agenda

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nheuillet/calendar-linker/intra"
	"github.com/nheuillet/calendar-linker/parser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
}

//GetEvents lists the next 250 events in the calendar
func GetEvents(srv *calendar.Service, agenda string) *calendar.Events {

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(agenda).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(250).OrderBy("startTime").Do()
	if err != nil {
		log.Printf("Unable to retrieve next ten of the user's events: %v\n", err)
	}
	return events
}

// parseTime is a very ugly piece of code because Epitech returned time is not golang time compliant (unless i'm an idiot)
// I'm using time.now().format(rfc.RFC3339 to get the local timezone), while
// building manually the rfc3339 timestamp. It's ugly and I'm ashamed to do it but it works for now.....
func parseTime(timeStr string) string {
	tmpTime := time.Now().Format(time.RFC3339)

	timeStr = strings.Replace(timeStr, " ", "T", 1)
	tmpPieceTimezone := strings.Split(tmpTime, ":")[2:]
	timeStr += tmpPieceTimezone[0][2:] + ":" + tmpPieceTimezone[1]

	return timeStr
}

func getTime(ev intra.Event) (string, string) {
	var start string
	end := parseTime(ev.End)

	if ev.RdvGroupRegistered != "" {
		start = parseTime(strings.Split(ev.RdvGroupRegistered, "|")[0])
		end = parseTime(strings.Split(ev.RdvGroupRegistered, "|")[1])
	} else if ev.RdvIndivRegistered != "" {
		start = parseTime(strings.Split(ev.RdvIndivRegistered, "|")[0])
		end = parseTime(strings.Split(ev.RdvIndivRegistered, "|")[1])
	} else {
		start = parseTime(ev.Start)
	}

	return start, end
}

func getAttendees(config *parser.Config, ev *intra.Activity) *[]*calendar.EventAttendee {
	var attendees []*calendar.EventAttendee

	if !config.ProjectParticipant {
		return nil
	}

	for index, at := range ev.Participants {
		attendees = append(attendees, &calendar.EventAttendee{
			Email:       at,
			DisplayName: ev.ParticipantsName[index],
		})
	}
	return &attendees
}

func createProjects(srv *calendar.Service, config *parser.Config, projects *[]intra.Activity) {
	calEvents := GetEvents(srv, config.GoogleCalendarProjects)
	for _, cEv := range calEvents.Items {
		for index, ev := range *projects {
			if cEv.Summary == ev.Title { // on create event the acti code is written in the description.
				if config.ProjectParticipant && ev.Participants != nil && cEv.Attendees == nil {
					//Case where the project was already created before but now the
					//Project started and the group has been created.
					//This way the group is added to the event, provided the
					//option is enabled in the config.
					(*projects)[index].Update = true
					(*projects)[index].ID = cEv.Id
					break
				}
				if index < len(*projects) {
					(*projects)[index] = (*projects)[len(*projects)-1]
				}
				*projects = (*projects)[:len(*projects)-1]
				break
			}
		}
	}
	for _, ev := range *projects {
		start := parseTime(ev.Begin)
		end := parseTime(ev.End)
		newEvent := &calendar.Event{
			Summary:     ev.Title,
			Description: ev.Description,
			Start: &calendar.EventDateTime{
				DateTime: start,
				TimeZone: config.Timezone,
			},
			End: &calendar.EventDateTime{
				DateTime: end,
				TimeZone: config.Timezone,
			},
			ColorId:   config.ProjectColor,
			Attendees: *getAttendees(config, &ev),
		}
		if ev.Update {
			_, err := srv.Events.Update(config.GoogleCalendarProjects, ev.ID, newEvent).Do()
			if err != nil {
				log.Printf("Unable to update event. %v\n", err)
			}
		} else {
			_, err := srv.Events.Insert(config.GoogleCalendarProjects, newEvent).Do()
			if err != nil {
				log.Printf("Unable to create event. %v\n", err)
			}
		}
	}
}

//CreateEvents creates the events passed on the callendar specified
func CreateEvents(srv *calendar.Service, config *parser.Config, events *[]intra.Event, projects *[]intra.Activity) {
	calEvents := GetEvents(srv, config.GoogleCalendarEvents)
	eventReminders := []*calendar.EventReminder{}

	for _, cEv := range calEvents.Items {
		for index, ev := range *events {
			if cEv.Description == ev.CodeActi { // on create event the acti code is written in the description.
				if index < len(*events) {
					(*events)[index] = (*events)[len(*events)-1]
				}
				*events = (*events)[:len(*events)-1]
			}
		}
	}

	for _, val := range config.Reminders {
		newReminder := calendar.EventReminder{
			Method:  "popup",
			Minutes: int64(val),
		}
		eventReminders = append(eventReminders, &newReminder)
	}
	for _, ev := range *events {
		start, end := getTime(ev)
		newEvent := &calendar.Event{
			Summary:     ev.ActiTitle,
			Location:    ev.Room.Code,
			Description: ev.CodeActi,
			Start: &calendar.EventDateTime{
				DateTime: start,
				TimeZone: config.Timezone,
			},
			End: &calendar.EventDateTime{
				DateTime: end,
				TimeZone: config.Timezone,
			},
			ColorId: config.EventColor,
			Reminders: &calendar.EventReminders{
				Overrides: eventReminders,
			},
		}
		newEvent.Reminders.ForceSendFields = []string{"UseDefault"}
		_, err := srv.Events.Insert(config.GoogleCalendarEvents, newEvent).Do()
		if err != nil {
			log.Printf("Unable to create event. %v\n", err)
		}
	}
	if projects != nil {
		createProjects(srv, config, projects)
	}
}

// GetGoogleClient initialize a client in order to see and create calendar events
func GetGoogleClient() *calendar.Service {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return srv
}
