package main

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// The helpers build image payloads at runtime so the tests carry no binary
// fixtures and never reach a network.

func pngPayload(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jpegPayload(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4)), nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestHashBytes(t *testing.T) {
	// sha1 of "test-data" in base64 url encoding, precomputed.
	if got := hashBytes([]byte("test-data")); got != "cRXpiQ9bXMaRS9-jt8AR2xza_ts=" {
		t.Error("Hash was not correct, hash was " + got)
	}
}

func TestTestImg(t *testing.T) {
	if !testImg(pngPayload(t)) {
		t.Error("testImg rejected a valid png")
	}
	if !testImg(jpegPayload(t)) {
		t.Error("testImg rejected a valid jpeg")
	}
	if testImg([]byte("<html><body>not an image</body></html>")) {
		t.Error("testImg accepted an html payload as an image")
	}
	if testImg(nil) {
		t.Error("testImg accepted an empty payload as an image")
	}
}

func TestNormalizeURL(t *testing.T) {
	if got := normalizeURL("//images.igdb.com/x.png"); got != "https://images.igdb.com/x.png" {
		t.Error("protocol relative url was not normalized, got " + got)
	}
	if got := normalizeURL("https://example.com/x.png"); got != "https://example.com/x.png" {
		t.Error("absolute url was changed, got " + got)
	}
}

//probeStub is a getImage probe that validates a fixed url set and records the
//urls it was asked about, so tests can assert the choice and that probing
//stopped as early as the code promises.
type probeStub struct {
	valid  map[string]bool
	probed []string
}

func newProbe(urls ...string) *probeStub {
	p := &probeStub{valid: make(map[string]bool)}
	for _, u := range urls {
		p.valid[u] = true
	}
	return p
}

func (p *probeStub) probe(URL string) bool {
	p.probed = append(p.probed, URL)
	return p.valid[URL]
}

func TestGetImgWithImg(t *testing.T) {
	//No name match: the first cover that passes the probe is used, and its
	//protocol relative url comes back normalized.
	mockGames := []igdbGame{
		{Name: "some other game", Cover: igdbCover{URL: "//test.local/a.png"}},
	}
	URL := getImage(mockGames, "game1", newProbe("https://test.local/a.png").probe)
	if URL != "https://test.local/a.png" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}

func TestGetImgWithExactName(t *testing.T) {
	mockGames := []igdbGame{
		{Name: "some other game", Cover: igdbCover{URL: "//test.local/a.png"}},
		{Name: "wanted game", Cover: igdbCover{URL: "//test.local/b.png"}},
		{Name: "never probed", Cover: igdbCover{URL: "//test.local/c.png"}},
	}
	probe := newProbe("https://test.local/a.png", "https://test.local/b.png", "https://test.local/c.png")
	URL := getImage(mockGames, "wanted game", probe.probe)
	if URL != "https://test.local/b.png" {
		t.Errorf("exact name did not win, URL was " + URL)
	}
	if len(probe.probed) != 2 || probe.probed[1] != "https://test.local/b.png" {
		t.Errorf("probing should stop at the exact match, probed %v", probe.probed)
	}
}

func TestGetImgWithDeadProbe(t *testing.T) {
	//Every candidate fails the probe: getImage reports nothing, which the
	//caller turns into the placeholder icon.
	mockGames := []igdbGame{
		{Name: "game1", Cover: igdbCover{URL: "//test.local/a.png"}},
		{Name: "game2", Cover: igdbCover{URL: "//test.local/b.png"}},
	}
	if URL := getImage(mockGames, "wanted", newProbe().probe); URL != "" {
		t.Errorf("dead covers should yield no url, got " + URL)
	}
}

func TestGetImgSkipsCoverlessGames(t *testing.T) {
	//The exact name has no cover, so it must not stop the scan or be probed.
	mockGames := []igdbGame{
		{Name: "wanted", Cover: igdbCover{}},
		{Name: "some other game", Cover: igdbCover{URL: "//test.local/a.png"}},
	}
	probe := newProbe("https://test.local/a.png")
	if URL := getImage(mockGames, "wanted", probe.probe); URL != "https://test.local/a.png" {
		t.Errorf("URL was " + URL)
	}
	if len(probe.probed) != 1 {
		t.Errorf("coverless game should not be probed, probed %v", probe.probed)
	}
}

func TestGetImgWithNo(t *testing.T) {
	mockGames := []igdbGame{}
	URL := getImage(mockGames, "game1", newProbe().probe)
	if URL != "" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}

//requireMongo returns a connected client or skips the test when no local
//database is running, so go test passes on a bare machine while the dev
//compose stack (which runs mongo) still gets full coverage. 127.0.0.1 is
//spelled out because the old driver does not fall back when localhost
//resolves to IPv6 first and the port is bound IPv4-only.
func requireMongo(t *testing.T) *mongo.Client {
	t.Helper()
	uri := os.Getenv("DB_STRING")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Skip("could not reach " + uri + ": " + err.Error())
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(context.Background())
		t.Skip("mongodb not running at " + uri + ": podman run -d -p 27017:27017 docker.io/library/mongo:6.0")
	}
	return client
}

func TestImageSetup(t *testing.T) {
	client := requireMongo(t)
	defer client.Disconnect(context.Background())

	//The cached icon lives on a local test server and the seeded hash matches
	//its bytes, so setup() exercises the cache-hit path without IGDB.
	iconBytes := pngPayload(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/icon.png" {
			w.Write(iconBytes)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	currentDir, _ := os.Getwd()
	image := imageGenerate{
		dir:        path.Join(t.TempDir(), "img"),
		staticDir:  path.Join(currentDir, "genImage", "Static"),
		profileURL: "https://cdn.discordapp.com/embed/avatars/0.png",
		userID:     "123",
		name:       "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//Adds the collections to the global vars so the program can access them
	statsCollection = client.Database("testing").Collection("gamestats")
	iconCollection = client.Database("testing").Collection("gameicons")
	settingCollection = client.Database("testing").Collection("settings")
	defer statsCollection.Drop(ctx)
	defer iconCollection.Drop(ctx)
	defer settingCollection.Drop(ctx)

	if _, err := statsCollection.InsertOne(ctx, stat{ID: "123", Game: "game1", Hours: 1, Ignore: false}); err != nil {
		t.Fatal("Error inserting stat " + err.Error())
	}
	if _, err := iconCollection.InsertOne(ctx, icon{Name: "game1", URL: server.URL + "/icon.png", Hash: hashBytes(iconBytes)}); err != nil {
		t.Fatal("Error inserting icon " + err.Error())
	}

	if _, err := image.setup(); err != nil {
		t.Fatal("Error setting up image " + err.Error())
	}

	for _, f := range []string{"data", path.Join("game1", "hours"), path.Join("game1", "icon")} {
		if _, err := os.Stat(path.Join(image.dir, f)); err != nil {
			t.Error("Error reading " + f + " " + err.Error())
		}
	}
	//The saved icon must be the served bytes, not the placeholder.
	saved, err := ioutil.ReadFile(path.Join(image.dir, "game1", "icon"))
	if err != nil || !bytes.Equal(saved, iconBytes) {
		t.Error("icon file does not match the served cover")
	}
}
