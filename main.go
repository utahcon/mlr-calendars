package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/utahcon/mlr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"time"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
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
	json.NewEncoder(f).Encode(token)
}

func main() {

	ctx := context.Background()
	b, err := os.ReadFile("mlr-calendar-app-creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	//Delete all calendars
	//for calName, calendarId := range mlr.CalendarLink {
	//	log.Println("Getting events on calendar:", calName, calendarId)
	//	events, err := srv.Events.List(calendarId).TimeMin(time.Now().Format(time.RFC3339)).Do()
	//	if err != nil {
	//		log.Fatalf("Error getting event list for %s: %v", calendarId, err)
	//	}
	//
	//	for _, event := range events.Items {
	//		attempts := 0
	//		for {
	//			attempts++
	//			log.Println("Deleting: ", calendarId, event.Id)
	//			err := srv.Events.Delete(calendarId, event.Id).Do()
	//			if err != nil {
	//				log.Printf("Error deleting event (sleep: %d): %v", time.Second*10*time.Duration(attempts), err)
	//				time.Sleep(time.Second * 10 * time.Duration(attempts))
	//			} else {
	//				break
	//			}
	//		}
	//	}
	//}

	mlrSrv, err := mlr.NewService()
	if err != nil {
		log.Fatalf("Error getting MLR Client: %v", err)
	}

	continuePolling := true
	page := 0
	for {
		if !continuePolling {
			break
		}
		page++

		log.Println("Polling Page: ", page)
		matchList, err := mlrSrv.Matches.List().TeamOne("").TeamTwo("").Count(10).Page(page).ExcludePlayers(true).Series("").Season("").Do()
		if err != nil {
			log.Fatalf("Error getting matches: %v", err)
		}

		if len(matchList.Matches) == 0 {
			continuePolling = false
		}

		for _, match := range matchList.Matches {

			// Create events on Calendar
			event := calendar.Event{
				Start: &calendar.EventDateTime{
					DateTime: match.Date.Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: match.Date.Add(time.Hour * 2).Format(time.RFC3339),
				},
				Location:    match.Venue.Name,
				Summary:     fmt.Sprintf("%s @ %s", match.AwayTeam.Name, match.HomeTeam.Name),
				Description: fmt.Sprintf("%s take on %s at %s\n\nBroadcasters: (not implemented)", match.HomeTeam.Name, match.AwayTeam.Name, match.Venue.Name),
				ExtendedProperties: &calendar.EventExtendedProperties{
					Private: map[string]string{"GUID": match.GUID},
				},
			}

			// Make sure Home Team is valid
			if _, ok := mlr.CalendarLink[match.HomeTeamId]; ok {
				event.Attendees = append(event.Attendees, &calendar.EventAttendee{
					ResponseStatus: "accepted",
					Email:          mlr.CalendarLink[match.HomeTeamId],
				})
			}

			// Make sure Away Team if valid
			if _, ok := mlr.CalendarLink[match.AwayTeamId]; ok {
				event.Attendees = append(event.Attendees, &calendar.EventAttendee{
					ResponseStatus: "accepted",
					Email:          mlr.CalendarLink[match.AwayTeamId],
				})
			}

			//MLR Main Calendar
			events, err := srv.Events.List(mlr.CalendarLink["MajorLeagueRugby"]).PrivateExtendedProperty(fmt.Sprintf("GUID=%s", match.GUID)).Do()
			if err != nil {
				log.Fatalf("Error getting list of matching events: %v", err)
			}

			if len(events.Items) >= 1 && events.Items[0].ExtendedProperties.Private["GUID"] == match.GUID {
				_, err := srv.Events.Update(mlr.CalendarLink["MajorLeagueRugby"], events.Items[0].Id, &event).Do()
				if err != nil {
					log.Fatalf("Error updating event %s: %v", events.Items[0].Id, err)
				}
			} else {
				_, err := srv.Events.Insert(mlr.CalendarLink["MajorLeagueRugby"], &event).Do()
				if err != nil {
					log.Fatalf("Error inserting event: %v", err)
				}
			}
		}
	}
}
