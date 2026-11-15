//Searches IGDB for a box art image matching the gameName.
//IGDB authenticates with a Twitch app access token obtained via the
//client credentials flow, cached in memory until close to expiry.

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type igdbCover struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type igdbGame struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Cover igdbCover `json:"cover"`
}

type twitchToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

var (
	tokenMutex   sync.Mutex
	cachedToken  string
	tokenExpires time.Time
)

//Endpoints are vars so tests can point them at local fakes.
var (
	twitchTokenURL = "https://id.twitch.tv/oauth2/token"
	igdbAPIURL     = "https://api.igdb.com/v4/games"
)

//apqlEscaper makes a game name safe to embed in a quoted ApQL string.
var apqlEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

//twitchAccessToken returns a cached Twitch app access token, fetching a new
//one when it is missing, close to expiring, or a refresh was forced.
func twitchAccessToken(force bool) (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if !force && cachedToken != "" && time.Now().Before(tokenExpires) {
		return cachedToken, nil
	}

	form := url.Values{}
	form.Set("client_id", os.Getenv("TWITCH_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("TWITCH_CLIENT_SECRET"))
	form.Set("grant_type", "client_credentials")

	client := new(http.Client)
	client.Timeout = time.Second * 10
	resp, err := client.PostForm(twitchTokenURL, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("Twitch token status code was " + strconv.Itoa(resp.StatusCode) + "\n")
	}

	//Parses the JSON from Twitch
	body, _ := ioutil.ReadAll(resp.Body)
	var token twitchToken
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", err
	}

	//Twitch tokens last about 60 days, refresh a day early to be safe.
	cachedToken = token.AccessToken
	tokenExpires = time.Now().Add(time.Duration(token.ExpiresIn-86400) * time.Second)
	return cachedToken, nil
}

//igdbQuery posts an ApQL query against the IGDB games endpoint, refreshing a
//stale Twitch token once on 401, and decodes the results.
func igdbQuery(query string) ([]igdbGame, error) {
	client := new(http.Client)
	client.Timeout = time.Second * 10

	//IGDB answers with 401 if the cached token went stale server side, the
	//second attempt then forces a token refresh.
	var body []byte
	for attempt := 0; ; attempt++ {
		token, err := twitchAccessToken(attempt > 0)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", igdbAPIURL, strings.NewReader(query))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Client-ID", os.Getenv("TWITCH_CLIENT_ID"))
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 401 && attempt == 0 {
			continue
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("IGDB Status Code Was " + strconv.Itoa(resp.StatusCode) + "\n")
		}
		break
	}

	//Parses the JSON from IGDB
	var searchResults []igdbGame
	err := json.Unmarshal(body, &searchResults)
	if err != nil {
		return nil, err
	}
	return searchResults, nil
}

//Searches IGDB for games matching the gameName.
//A deterministic exact-name query runs first: `where name = "X"` can only
//return name matches, so DLC and season entries are impossible. Only when it
//comes back empty (casing or spelling differences, e.g. "DOOM" vs "Doom")
//does the fuzzy search run; its ranking favours recent DLC over the base
//game, so the result window is wide enough for getImage to exact-match in.
func search(gameName string) ([]igdbGame, error) {
	name := apqlEscaper.Replace(gameName)
	exact := "fields name, cover.url; where name = \"" + name + "\"; limit 5;"
	games, err := igdbQuery(exact)
	if err != nil {
		return nil, err
	}
	if len(games) > 0 {
		return games, nil
	}
	fuzzy := "fields name, cover.url; search \"" + name + "\"; limit 10;"
	return igdbQuery(fuzzy)
}
