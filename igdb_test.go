package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

//fakeIGDB is a local stand-in for Twitch + IGDB. Handlers appended to
//Queries and counted TokenCalls through the struct the helper hands them.
type fakeIGDB struct {
	Queries    []string
	TokenCalls int
}

//useFakeServices points the Twitch token endpoint and the IGDB games endpoint
//at local test servers. The handler answers the games queries and tests
//assert on the captured ApQL. Token and token cache state is reset around
//each test.
func useFakeServices(t *testing.T, handler func(f *fakeIGDB, w http.ResponseWriter, r *http.Request)) *fakeIGDB {
	t.Helper()
	f := &fakeIGDB{}

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.TokenCalls++
		fmt.Fprint(w, `{"access_token":"tok","expires_in":999999,"token_type":"bearer"}`)
	}))
	gamesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		f.Queries = append(f.Queries, string(body))
		handler(f, w, r)
	}))

	oldToken, oldAPI := twitchTokenURL, igdbAPIURL
	twitchTokenURL, igdbAPIURL = tokenServer.URL, gamesServer.URL+"/games"
	cachedToken, tokenExpires = "", time.Time{}
	t.Cleanup(func() {
		tokenServer.Close()
		gamesServer.Close()
		twitchTokenURL, igdbAPIURL = oldToken, oldAPI
		cachedToken, tokenExpires = "", time.Time{}
	})
	return f
}

func respond(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, body)
}

const fortniteHit = `[{"name":"Fortnite","cover":{"url":"//cdn/f.png"}}]`
const doomHit = `[{"name":"Doom","cover":{"url":"//cdn/d.png"}}]`

func TestSearchUsesExactTierFirst(t *testing.T) {
	f := useFakeServices(t, func(f *fakeIGDB, w http.ResponseWriter, r *http.Request) {
		respond(w, fortniteHit)
	})

	games, err := search("Fortnite")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].Name != "Fortnite" {
		t.Fatalf("unexpected results: %+v", games)
	}
	if len(f.Queries) != 1 {
		t.Fatalf("a hit on the exact tier must not run the fuzzy query, got %d queries: %v", len(f.Queries), f.Queries)
	}
	if !strings.Contains(f.Queries[0], `where name = "Fortnite"`) {
		t.Errorf("first query should be the exact one, got %s", f.Queries[0])
	}
	if f.TokenCalls != 1 {
		t.Errorf("token should be fetched once, got %d", f.TokenCalls)
	}
}

func TestSearchFallsBackToFuzzy(t *testing.T) {
	f := useFakeServices(t, func(f *fakeIGDB, w http.ResponseWriter, r *http.Request) {
		query := f.Queries[len(f.Queries)-1]
		if strings.Contains(query, "search") {
			respond(w, doomHit)
		} else {
			respond(w, `[]`) //the exact tier misses on DOOM vs Doom
		}
	})

	games, err := search("DOOM")
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Queries) != 2 {
		t.Fatalf("expected exact then fuzzy, got %v", f.Queries)
	}
	if !strings.Contains(f.Queries[1], `search "DOOM"`) || !strings.Contains(f.Queries[1], "limit 10") {
		t.Errorf("fuzzy query wrong: %s", f.Queries[1])
	}
	if len(games) != 1 || games[0].Name != "Doom" {
		t.Errorf("unexpected results: %+v", games)
	}
}

func TestSearchEscapesApQL(t *testing.T) {
	f := useFakeServices(t, func(f *fakeIGDB, w http.ResponseWriter, r *http.Request) {
		respond(w, `[]`)
	})

	if _, err := search(`Owe"n\`); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Queries, "\n")
	if strings.Contains(joined, `Owe"n`) {
		t.Errorf("quote reached the query unescaped: %s", joined)
	}
	if !strings.Contains(joined, `Owe\"n\\`) {
		t.Errorf("escaped name not found in queries: %s", joined)
	}
}

func TestSearchRefreshesTokenOn401(t *testing.T) {
	attempts := 0
	f := useFakeServices(t, func(f *fakeIGDB, w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		respond(w, fortniteHit)
	})

	if _, err := search("Fortnite"); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Errorf("expected one retry after 401, attempts = %d", attempts)
	}
	if f.TokenCalls != 2 {
		t.Errorf("a forced token refresh was expected after the 401, token fetches = %d", f.TokenCalls)
	}
}

func TestTokenIsCached(t *testing.T) {
	f := useFakeServices(t, func(f *fakeIGDB, w http.ResponseWriter, r *http.Request) {
		respond(w, fortniteHit)
	})

	if _, err := search("Fortnite"); err != nil {
		t.Fatal(err)
	}
	if _, err := search("Overwatch"); err != nil {
		t.Fatal(err)
	}
	if f.TokenCalls != 1 {
		t.Errorf("token should be cached between searches, got %d fetches", f.TokenCalls)
	}
}
