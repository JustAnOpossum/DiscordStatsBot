//Searches and finds game icons and saves them to the database

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type bingAnswer struct {
	Type            string `json:"_type"`
	Instrumentation struct {
		PageLoadPingURL interface{} `json:"pageLoadPingUrl"`
	} `json:"instrumentation"`
	WebSearchURL          string      `json:"webSearchUrl"`
	TotalEstimatedMatches int         `json:"totalEstimatedMatches"`
	Value                 []bingValue `json:"value"`
	QueryExpansions       []struct {
		Text         string      `json:"text"`
		DisplayText  string      `json:"displayText"`
		WebSearchURL string      `json:"webSearchUrl"`
		SearchLink   string      `json:"searchLink"`
		Thumbnail1   interface{} `json:"thumbnail1"`
	} `json:"queryExpansions"`
	NextOffsetAddCount int `json:"nextOffsetAddCount"`
	PivotSuggestions   []struct {
		Pivot       string `json:"pivot"`
		Suggestions []struct {
			Text         string `json:"text"`
			DisplayText  string `json:"displayText"`
			WebSearchURL string `json:"webSearchUrl"`
			SearchLink   string `json:"searchLink"`
			Thumbnail    struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"thumbnail"`
		} `json:"suggestions"`
	} `json:"pivotSuggestions"`
	DisplayShoppingSourcesBadges bool        `json:"displayShoppingSourcesBadges"`
	DisplayRecipeSourcesBadges   bool        `json:"displayRecipeSourcesBadges"`
	SimilarTerms                 interface{} `json:"similarTerms"`
}

type bingValue struct {
	Name               string      `json:"name"`
	DatePublished      string      `json:"datePublished"`
	HomePageURL        interface{} `json:"homePageUrl"`
	ContentSize        string      `json:"contentSize"`
	HostPageDisplayURL string      `json:"hostPageDisplayUrl"`
	Width              int         `json:"width"`
	Height             int         `json:"height"`
	Thumbnail          struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"thumbnail"`
	ImageInsightsToken     string      `json:"imageInsightsToken"`
	InsightsSourcesSummary interface{} `json:"insightsSourcesSummary"`
	ImageID                string      `json:"imageId"`
	AccentColor            string      `json:"accentColor"`
	WebSearchURL           string      `json:"webSearchUrl"`
	ThumbnailURL           string      `json:"thumbnailUrl"`
	EncodingFormat         string      `json:"encodingFormat"`
	ContentURL             string      `json:"contentUrl"`
}

//Searches bing for an icon mathing the gameName
func search(gameName string) (bingAnswer, error) {
	//Starts by creating a new requst to amke the bing search
	query := url.QueryEscape(gameName + " icon")
	req, _ := http.NewRequest("GET", "https://api.cognitive.microsoft.com/bing/v7.0/images/search?q="+query, nil)
	req.Header.Add("Ocp-Apim-Subscription-Key", os.Getenv("BING_KEY"))
	client := new(http.Client)
	client.Timeout = time.Second * 5

	//Client does the bing search request
	resp, err := client.Do(req)
	if err != nil {
		return bingAnswer{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return bingAnswer{}, errors.New("Bing Status Code Was " + strconv.Itoa(resp.StatusCode) + "\n")
	}

	//Parses the JSON from bing
	body, _ := ioutil.ReadAll(resp.Body)
	var searchResults bingAnswer
	err = json.Unmarshal(body, &searchResults)
	if err != nil {
		return bingAnswer{}, err
	}
	return searchResults, nil
}
