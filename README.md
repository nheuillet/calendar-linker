# calendar-linker

[![Go Report Card](https://goreportcard.com/badge/github.com/nheuillet/calendar-linker)](https://goreportcard.com/report/github.com/nheuillet/calendar-linker)

An **easy** and **fast** linker that adds your Epitech calendar to your Google / Outlook (W.I.P) calendar.


# What you **can** do

- Sync Projects and events to your chosen calendar.
    - Correct hours (appointments too! if your appointments is 9:30 -> 9:45, the same goes for your created event on your calendar)
    - Group sync: once your group is created for a project it will be added on your calendar, whether the project was already created or not.
    - ClassRoom added as location of the created event
    - Project description added as event description if there is one
- Customizable colors for Events and / or Projects
- As much reminders as you want: do you want to get a notification 30 and 10 minutes before every event? Sure thing. Notifications are not your thing? No problem, just leave this option empty, it is now disabled.
- Disable / Enable project creation and / or group introspection whenever you want (note: project+ members introspection adds between 5 and 15 seconds for the program to finish, use it wisely!)
- Custom Regex in order to extract a room code from the intranet answer. Done this way so that it works with every epitech cities.

# What you **can't** do (yet?)

- Specify different color according to modules the event is from
- Microsoft Outlook Calendar sync (although you can sync your google calendar to your outlook calendar, so that the result is almost the same)


# Installation:

## Basics
- Clone this repository.
- Run:
```bash
    go build
```
- Modify the `config.json` file with the following content:

```json
{
    "google_calendar_events": "XXXX@group.calendar.google.com",
    "google_calendar_projects": "XXXX@group.calendar.google.com",
    "epitech_auth": "auth-XXXXXXXX",
    "epitech_location_code": "FR/TLS",
    "create_project_event": false,
    "add_participants_to_project": true,
    "epitech_semesters": [5, 6],
    "timezone": "Europe/Paris",
    "event_color": "10",
    "project_color": "3",
    "reminder_time": [10, 30],
    "location_regex": "\\w{2}\/\\w+\/\\w+\/([\\w-_]+)"
}

```

| field | explanation | notes |
|-------|-------------|-------|
|google_calendar_events|The google calendar ID where to create the daily events| Mandatory
|google_calendar_projects|he google calendar ID where to create the projects events.| Creating a different calendar is highly recommended or your calendar will be unreadable, but using the same as `google_calendar_events` is also possible.|
|epitech_auth| Epitech Autologin link | Found [here](https://intra.epitech.eu/admin/autolog)|
|epitech_location_code| the epitech location code | Example: `FR/TLS` for Toulouse. Used in order to filter events on search|
|create_project_event|Turning to true enable the creation of the Projects events on the calendar|Default is `false`. The slowest option (adds between 5 and 15 seconds during tests depending of the number of semesters you choose)|
|add_participants_to_project|Adds the participants to the projects. Works even if the project was already created. |Default is `false`. Adds N * Requests to the intra api, N being the number of projects found in the semester list. As every |
|epitech_semesters| The semesters you want to scan if create_project_event is set to True||
|timezone | the timezone of the Epitech School | Default is `Europe/Paris`
|event_color|The color of the google calendar event created for the daily events|The google calendar color code. Default is  `"3"` (string). See [here](https://lukeboyle.com/blog-posts/2016/04/google-calendar-api---color-id) for references.
|project_color|The color of the google calendar event created for the projects|The google calendar color code. Default is  `"10"` (string). See [here](https://lukeboyle.com/blog-posts/2016/04/google-calendar-api---color-id) for references.
|reminder_time|Array of minutes for the google calendar reminders|Default is `[10, 30]`, but it can be very annoying to get 2 notifications for each class. Leave it empty (`[]`) for no notifications
|location_regex|The regex used to clean the room name| Default is `\\w{2}\/\\w+\/\\w+\/([\\w-_]+)`. Example of room to be cleaned: `FR/TLS/Marquette/703` will lead to `703`


## Configuration

### Google Calendar

- First, it is highly recommended to create 2 sub-calendars: one for daily classes and one for the projects events, as those tend to make your normal calendar unreadable.

![New Calendar](https://i.imgur.com/THXXkR0.png)

Then go to `Settings and sharing` menu option for this sub-calendar:

![Settings and sharing](https://i.imgur.com/mvFQdWB.jpg)

Go to `Calendar ID` and copy-paste the address `XXXX@group.calendar.google.com` in the `config.json` file for the correct key, according to the table above.


### APIs configuration

Go to [The Calendar Api Documentation](https://developers.google.com/calendar/quickstart/go) and click on "Enable the Google Calendar API". <br>
Create a project and download the client configuration. <br>
Put the `credentials.json` in the program directory, alongside `config.json`.


run ./calendar-linker to execute the program. <br>
*You will need to connect to your Google account the first time.*

# Disclaimer

Code is not the best. This project was made more of a POC because of frustration than anything else. While it was **way** worse at the beginning, there are still plenty of room for improvements. A lot of things are ugly workarounds in order to achieve a result in the fastest/easiest way, as I didn't spend nearly enough hours on this code to make it clean. Also, Epitech's intranet is full of bad practices that requires to create even more workarounds (Looking at you `registered` field that is either a string or a bool - `"registered"` or `false`. Why...).

My Code is definitely not the cleanest or the fastest. <br>
Feel free to criticize me and my code as much as you want, provided that you show me how to do better.

# Question? Feature request?

Feel free to open an issue. I'll try my best to answer.

# Notes

- Use a crontab to automatically sync your Epitech calendar regularly. See [Crontab Guru for examples](https://crontab.guru).

- This project was **heavily** inspired by the [Epitech_To_Google_Calendar](https://github.com/Thezap/Linker_EPITECH_To_GOOGLE_Calendar) project. I really didn't like how **slow** it was + a few other things, so I decided to rewrite my own, in golang :heart:.
Some images used are the exact same as this project. Again, Credits to @Thezap.
