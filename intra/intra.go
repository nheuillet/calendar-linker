package intra

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nheuillet/calendar-linker/parser"
)

// Event is a json structure of the elements in the intra json response
type Event struct {
	CodeModule         string      `json:"codemodule"`
	CodeInstance       string      `json:"codeinstance"`
	CodeActi           string      `json:"codeacti"`
	CodeEvent          string      `json:"codeevent"`
	ModuleTitle        string      `json:"titlemodule"`
	ActiTitle          string      `json:"acti_title"`
	Start              string      `json:"start"`
	End                string      `json:"end"`
	IsRdv              string      `json:"is_rdv"`
	Room               Room        `json:"room"`
	RawEventRegistered interface{} `json:"event_registered"`
	EventRegistered    bool        `json:"-"`
	RdvGroupRegistered string      `json:"rdv_group_registered"`
	RdvIndivRegistered string      `json:"rdv_indiv_registered"`
}

// Module is a structure that represents the intra Module
type Module struct {
	ID           int    `json:"id"`
	Semester     int    `json:"semester"`
	Scholaryear  int    `json:"scolaryear"`
	Code         string `json:"code"`
	Codeinstance string `json:"codeinstance"`
	Title        string `json:"title"`
	Registered   string `json:"status"`
}

// Activities is a struct containing an array of activies. Only there
// in order to match the intra output so that I can easily unmarshall to json
type Activities struct {
	Activities []Activity `json:"activites"`
}

// Activity is a struct that contains useful information for every activity
type Activity struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	Begin            string `json:"begin"`
	End              string `json:"end"`
	TypeTitle        string `json:"type_title"`
	IsProject        bool   `json:"is_projet"`
	CodeActi         string `json:"codeacti"`
	Participants     []string
	ParticipantsName []string
	Update           bool   // for calendar update purposes. I know it's ugly to put it here. Sorry
	ID               string // for calendar update purposes. I know it's ugly to put it here. Sorry
}

// Project is a struct that contains information about the project
type Project struct {
	Title         string       `json:"title"`
	UserGroupName string       `json:"user_project_title"`
	Registered    []Registered `json:"registered"`
}

//Member is a struct that contains information about the member of the group
type Member struct {
	Login string `json:"login"`
	Name  string `json:"title"`
}

// Registered is the list of groups
type Registered struct {
	Title   string   `json:"title"`
	Master  Member   `json:"master"`
	Members []Member `json:"members"`
}

// Room is a simple json object describing an epitech room field
type Room struct {
	Code  string `json:"code"`
	Type  string `json:"type"`
	Seats int    `json:"seats"`
}

const intraURL = "https://intra.epitech.eu/"
const intraTimeout = 10

// getHTTPClient spawns a net/http client and returns it.
func getHTTPClient() *http.Client {
	httpClient := &http.Client{Timeout: intraTimeout * time.Second}
	return httpClient
}

// getJSONResponse will execute a GET request then decode the answer as JSON.
func getJSONResponse(client *http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

//EpitechTimeToRFC allows to convert the Epitech Timestamp to RFC 3339
func EpitechTimeToRFC(timeStr string) string {
	tmp := strings.Split(timeStr, " ")

	newTime := tmp[0] + "T" + tmp[1] + "+01:00" // I know, it's hardCoded Location, sorry.

	return newTime
}

//inArray checks if the integer passed in the first parameter is in the int array in the second parameter
func inArray(val int, arr []int) bool {
	for _, value := range arr {
		if val == value {
			return true
		}
	}
	return false
}

//popModule removes pops the data at idx. CAUTION: replace by last data.
// You need to make sure the last data matches your condition.
func popModule(tab *[]Module, idx int) {
	if idx == len(*tab)-1 {
		*tab = (*tab)[:len(*tab)-1]
	} else {
		(*tab)[idx] = (*tab)[len(*tab)-1]
		*tab = (*tab)[:len(*tab)-1]
	}
}

//popActivity removes pops the data at idx. CAUTION: replace by last data.
// You need to make sure the last data matches your condition.
func popActivity(tab *[]Activity, idx int) {
	if idx == len(*tab)-1 {
		*tab = (*tab)[:len(*tab)-1]
	} else {
		(*tab)[idx] = (*tab)[len(*tab)-1]
		*tab = (*tab)[:len(*tab)-1]
	}
}

//popEvent removes pops the data at idx. CAUTION: replace by last data.
// You need to make sure the last data matches your condition.
func popEvent(tab *[]Event, idx int) {
	if idx == len(*tab)-1 {
		*tab = (*tab)[:len(*tab)-1]
	} else {
		(*tab)[idx] = (*tab)[len(*tab)-1]
		*tab = (*tab)[:len(*tab)-1]
	}
}

//cleanRoomName will clean the room name of the event according to the regex in the config file.
func cleanRoomName(events *[]Event, config *parser.Config) {
	reg := regexp.MustCompile(config.LocationRegex)
	for index, event := range *events {
		if (*events)[index].Room.Code == "" { // case where no room were provided in the activity
			continue
		}
		(*events)[index].Room.Code = reg.FindStringSubmatch(event.Room.Code)[1]
	}
}

// getCalendarRoute returns the calendar route appended with the auth token + the time
func getCalendarRoute(auth string, loc string) string {
	startTime := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	endTime := time.Now().AddDate(0, 2, 0).Format("2006-01-02")
	url := fmt.Sprintf("%s%s/planning/load?format=json&location=%s&onlymypromo=true&onlymymodule=true&start=",
		intraURL, auth, loc)
	url += fmt.Sprintf("%s&end=%s", startTime, endTime)

	return url
}

// GetRegisteredEvents fetches the list of events registered on a two month period starting from tomorrow.
func GetRegisteredEvents(conf *parser.Config, listEvents *[]Event) error {
	httpClient := getHTTPClient()
	url := getCalendarRoute(conf.EpitechAuth, conf.Location)
	err := getJSONResponse(httpClient, url, listEvents)
	if err != nil {
		return err
	}
	trimUnregisteredEvents(listEvents)
	err = trimFinishedEvents(listEvents)
	if err != nil {
		return err
	}
	cleanRoomName(listEvents, conf)
	return nil
}

// GetModules retrieves registered modules of the user
func GetModules(conf *parser.Config, client *http.Client) (*[]Module, error) {
	url := intraURL + conf.EpitechAuth + "/course/filter?format=json"
	modules := []Module{}

	err := getJSONResponse(client, url, &modules)
	if err != nil {
		return nil, err
	}
	trimUselessModules(&modules, conf)
	return &modules, nil
}

//GetProjects fetches the list of all the projects the user is registered to
// goroutines are used in order to improve the performances.
// After waiting for every module information to be sent to the channel, the said
// channel is read until all projects are retrieved.
func GetProjects(config *parser.Config, projects *[]Activity) error {
	chanProjects := make(chan []Activity, 75) // a buffered channel is used and set to 75 because it hangs if unbuffered. I'm not proficient enough with channels to know how to avoid this.
	client := getHTTPClient()
	modules, err := GetModules(config, client)
	var wg sync.WaitGroup

	if err != nil {
		return err
	}
	for _, module := range *modules {
		wg.Add(1)
		go GetModuleProjects(config, module, &chanProjects, &wg, client)
	}
	wg.Wait()
	close(chanProjects)
	for val := range chanProjects {
		*projects = append(*projects, val...)
	}
	trimEndedProjects(projects)
	return nil
}

// GetModuleProjects retrieves the projects for every modules
func GetModuleProjects(conf *parser.Config, module Module, chanProjects *chan []Activity, wg *sync.WaitGroup, client *http.Client) {
	url := intraURL + conf.EpitechAuth + "/module/"
	url += strconv.Itoa(module.Scholaryear) + "/" + module.Code + "/"
	url += module.Codeinstance
	projects := &Activities{}

	defer wg.Done()

	err := getJSONResponse(client, url+"/?format=json", projects)
	if err != nil {
		log.Println(err)
	}
	trimUselessActivities(projects)
	if len(projects.Activities) != 0 {
		if conf.ProjectParticipant {
			addProjectParticipant(client, projects, url)
		}
		*chanProjects <- projects.Activities
	}
}

//trimUselessModules will remove every semester not in the epitech_semester config field.
func trimUselessModules(modules *[]Module, conf *parser.Config) {
	for i := 0; i < len(*modules); i++ {
		if !inArray((*modules)[i].Semester, conf.Semesters) || (*modules)[i].Registered == "notregistered" {
			popModule(modules, i)
			i--
		}
	}
}

//trimUselessActivities will remove every activities that are not projects.
func trimUselessActivities(projects *Activities) {
	for i := 0; i < len(projects.Activities); i++ {
		if !projects.Activities[i].IsProject || (projects.Activities[i].TypeTitle != "Project" && projects.Activities[i].TypeTitle != "Mini-project") {
			popActivity(&projects.Activities, i) // Technically we don't need to check typetitle. Unfortunately, I've seen
			i--                                  // activities that are not projects or miniprojects but have  isProject set to true.
		}
	}
}

//addProjectParticipant will retrieve teammate list for every projects and append it to the project informations
func addProjectParticipant(client *http.Client, projects *Activities, url string) {
	for index, val := range projects.Activities {
		project := &Project{}
		err := getJSONResponse(client, url+"/"+val.CodeActi+"/project/?format=json", project)
		if err != nil {
			log.Println(err)
			return
		}
		for _, member := range project.Registered {
			if member.Title != project.UserGroupName {
				continue
			}
			projects.Activities[index].Participants = append(projects.Activities[index].Participants, member.Master.Login)
			projects.Activities[index].ParticipantsName = append(projects.Activities[index].ParticipantsName, member.Master.Name)
			for _, name := range member.Members {
				projects.Activities[index].Participants = append(projects.Activities[index].Participants, name.Login)
				projects.Activities[index].ParticipantsName = append(projects.Activities[index].ParticipantsName, name.Name)
			}
			break
		}
	}
}

// trimUnregisteredEvents removes events that we did not register to. those events are not supposed to
// exist as we specify only registered in the calendar, but some still get through.
func trimUnregisteredEvents(listEvents *[]Event) {
	i := 0
	for _, event := range *listEvents {
		if _, castOk := event.RawEventRegistered.(bool); castOk { // cast assertion for handling the field that is either False (bool) or "registered" (string). Yup. Intra api is bad.
			event.EventRegistered = false
		} else {
			event.EventRegistered = true
			(*listEvents)[i] = event
			i++
		}
	}
	*listEvents = (*listEvents)[:i]
}

//trimEndedProjects removes projects that ended before the program is run
func trimEndedProjects(projects *[]Activity) {
	now := time.Now()
	for i := 0; i < len(*projects); i++ {
		endTime, _ := time.Parse(time.RFC3339, EpitechTimeToRFC((*projects)[i].End))
		if now.After(endTime) {
			popActivity(projects, i)
			i--
		}
	}
}

//trimFinishedEvents removes event happening the same day the program is run but that already finished
func trimFinishedEvents(listEvents *[]Event) error {
	now := time.Now()
	for i := 0; i < len(*listEvents); i++ {
		t, err := time.Parse(time.RFC3339, EpitechTimeToRFC((*listEvents)[i].Start))
		if err != nil {
			return err
		}
		if now.After(t) {
			popEvent(listEvents, i)
			i-- // prevent itteration as we popped the current i
		}
	}
	return nil
}
