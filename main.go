package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/utahcon/mlr"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
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

	mlrSrv, err := mlr.NewService()
	if err != nil {
		log.Fatalf("Error getting MLR Client: %v", err)
	}

	matches, err := mlrSrv.Matches.List().TeamOne("").TeamTwo("").Count(10).Page(1).ExcludePlayers(true).Series("").Season("").Do()
	if err != nil {
		log.Fatalf("Error getting matches: %v", err)
	}
	for _, match := range matches.Matches {
		fmt.Println(match.GUID)
	}
	//
	//httpClient := &http.Client{}
	//
	//v := url.Values{}
	//v.Set("searchTeamOneId", "")
	//v.Set("searchTeamTwoId", "")
	//v.Set("seriesId", "")
	//v.Set("seasonId", "")
	//v.Set("pageSize", "5")
	//v.Set("pageNumber", "1")
	//v.Set("excludePlayers", "true")
	//
	//req, err := http.NewRequest("GET", mlr.SearchMatches+"?"+v.Encode(), nil)
	//
	//req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0")
	//req.Header.Add("Accept", "*/*")
	//req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	//req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	//req.Header.Add("Referer", "https://www.majorleague.rugby/")
	//req.Header.Add("Origin", "https://www.majorleague.rugby")
	//req.Header.Add("DNT", "1")
	//req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Sec-Fetch-Dest", "empty")
	//req.Header.Add("Sec-Fetch-Mode", "no-cors")
	//req.Header.Add("Sec-Fetch-Site", "cross-site")
	//req.Header.Add("Sec-GPC", "1")
	//req.Header.Add("TE", "trailers")
	//req.Header.Add("Content-Type", "application/json")
	//req.Header.Add("stratus-media-headers", "true")
	//req.Header.Add("EditOnBehalfOfUserGuid", "1EFF3C3F-1069-44FE-9B87-17F22B57597F")
	//req.Header.Add("tenantId", "4CFEC22F-C6D7-4BDC-A568-B1071DBD0B6E")
	//req.Header.Add("Pragma", "no-cache")
	//req.Header.Add("Cache-Control", "no-cache")
	//
	//resp, err := httpClient.Do(req)
	//if err != nil {
	//	log.Fatalf("Error calling MLR LiveScores: %v", err)
	//}
	//
	//fmt.Println("Response Headers:", resp.Header)
	//fmt.Println("Response Status Code:", resp.StatusCode)
	//fmt.Println("Response Status:", resp.Status)
	//body, err := io.ReadAll(resp.Body)
	//resp.Body.Close()
	//fmt.Println("Response:", string(body))
	//
	//var matches []mlr.Match
	//json.Unmarshal(body, &matches)
	//
	//for _, match := range matches {
	//	fmt.Println("Match Id:", match.GUID)
	//}

	//ctx := context.Background()
	//b, err := os.ReadFile("mlr-calendar-app-creds.json")
	//if err != nil {
	//	log.Fatalf("Unable to read client secret file: %v", err)
	//}
	//
	//// If modifying these scopes, delete your previously saved token.json.
	//config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	//if err != nil {
	//	log.Fatalf("Unable to parse client secret file to config: %v", err)
	//}
	//client := getClient(config)
	//
	//srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	//if err != nil {
	//	log.Fatalf("Unable to retrieve Calendar client: %v", err)
	//}
	//
	//cals, err := srv.CalendarList.List().ShowDeleted(false).Do()
	//for _, cal := range cals.Items {
	//	fmt.Printf("%s: %s\n", cal.Id, cal.Summary)
	//}
	//
	//t := time.Now().Format(time.RFC3339)
	//events, err := srv.Events.List("primary").ShowDeleted(false).SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	//if err != nil {
	//	log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	//}
	//fmt.Println("Upcoming events:")
	//if len(events.Items) == 0 {
	//	fmt.Println("No upcoming events found.")
	//} else {
	//	for _, item := range events.Items {
	//		date := item.Start.DateTime
	//		if date == "" {
	//			date = item.Start.Date
	//		}
	//		fmt.Printf("%v (%v)\n", item.Summary, date)
	//	}
	//}
}
