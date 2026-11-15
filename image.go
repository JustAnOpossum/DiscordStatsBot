//Methods for creating the image and downloading the game images
//Sets up the dir for images and creates the images, and removes the image
//Also has methods for downloading the images and finding the images from an IGDB search result.

package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type imageGenerate struct {
	dir        string
	staticDir  string
	profileURL string
	userID     string
	name       string
}

//Writes the bundled placeholder icon into the game dir. A game whose cover
//cannot be found or downloaded still renders (one flat tile) instead of
//failing the whole stats image, and nothing bad is cached, so the next
//request retries the real lookup.
func (img *imageGenerate) savePlaceholderIcon(imgDir string, gameName string) error {
	data, err := ioutil.ReadFile(path.Join(img.staticDir, "iconError.png"))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(imgDir, gameName, "icon"), data, 0744)
}

//Gets and saves an image, function so it can be called more than once
func (img *imageGenerate) getAndSaveImg(ctx context.Context, gameName string, imgDir string) error {
	var searchResults []igdbGame
	if results, err := search(gameName); err != nil {
		log.Println("IGDB search failed for " + gameName + ": " + err.Error())
	} else {
		searchResults = results
	}
	imgURL := getImage(searchResults, gameName, testImgURL)
	if imgURL == "" {
		log.Println("No IGDB cover found for " + gameName + ", using placeholder icon")
		return img.savePlaceholderIcon(imgDir, gameName)
	}
	body, err := download(imgURL)
	if err != nil {
		log.Println("Cover download failed for " + gameName + ": " + err.Error())
		return img.savePlaceholderIcon(imgDir, gameName)
	}
	//Only cache a cover that actually downloaded; the hash is computed from
	//the same bytes that get saved.
	_, err = iconCollection.InsertOne(ctx, bson.M{"name": gameName, "url": imgURL, "hash": hashBytes(body)})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(imgDir, gameName, "icon"), body, 0744)
}

//Sets up the dir for image generation
func (img *imageGenerate) setup() (string, error) {
	err := os.Mkdir(img.dir, 0744)
	if err != nil {
		return "", err
	}

	ctx, close := context.WithTimeout(context.Background(), time.Second*45)
	defer close()
	var limit int64 = 5
	var stats []stat
	cursor, _ := statsCollection.Find(ctx, bson.M{"id": img.userID, "ignore": false}, &options.FindOptions{
		Limit: &limit,
		Sort:  bson.M{"hours": -1},
	})
	cursor.All(ctx, &stats)

	var top5Games string
	//Loops though the stats to find all the game images
	for i := range stats {
		var userGame icon
		//Creates the dir to contain the game data

		//Fixes directory generation when a game has a / in it
		if strings.Contains(stats[i].Game, "/") {
			stats[i].Game = strings.ReplaceAll(stats[i].Game, "/", " ")
		}

		err = os.Mkdir(path.Join(img.dir, stats[i].Game), 0744)
		if err != nil {
			return "", err
		}
		//Writes the hours
		ioutil.WriteFile(path.Join(img.dir, stats[i].Game, "hours"), []byte(strconv.FormatFloat(stats[i].Hours, 'f', -1, 64)), 0744)
		err := iconCollection.FindOne(ctx, bson.M{"name": stats[i].Game}).Decode(&userGame)
		//Case if the game img does not exist
		if err == mongo.ErrNoDocuments {
			err = img.getAndSaveImg(ctx, stats[i].Game, img.dir)
			if err != nil {
				return "", err
			}
			//Case if it does exist
		} else {
			body, err := download(userGame.URL)
			if err != nil {
				//Transient failure keeps the cached document, so the next
				//request retries the download; the tile just uses the
				//placeholder this time.
				log.Println("Cached cover download failed for " + stats[i].Game + ": " + err.Error())
				if err := img.savePlaceholderIcon(img.dir, stats[i].Game); err != nil {
					return "", err
				}
			} else if hashBytes(body) != userGame.Hash {
				//The image behind the cached url changed, re-search for it.
				iconCollection.DeleteOne(ctx, bson.M{"name": userGame.Name})
				err = img.getAndSaveImg(ctx, stats[i].Game, img.dir)
				if err != nil {
					return "", err
				}
			} else {
				ioutil.WriteFile(path.Join(img.dir, stats[i].Game, "icon"), body, 0744)
			}
		}
		top5Games += stats[i].Game + "\n"
	}

	//Creates the main data file
	var hoursPlayed float64
	var gamesPlayed int
	var allStats []stat
	var settingsMenu setting
	cursor, err = statsCollection.Find(ctx, bson.M{"id": img.userID})
	if err != nil {
		return "", err
	}
	cursor.All(ctx, &allStats)
	for i := range allStats {
		hoursPlayed += allStats[i].Hours
	}
	distinct, err := statsCollection.Distinct(ctx, "game", bson.M{"id": img.userID})
	if err != nil {
		return "", err
	}
	gamesPlayed = len(distinct)
	settingCollection.FindOne(ctx, bson.M{"id": img.userID}).Decode(&settingsMenu)

	ioutil.WriteFile(path.Join(img.dir, "data"), []byte(strconv.Itoa(int(hoursPlayed))+"\n"+strconv.Itoa(gamesPlayed)+"\n"+img.name+"\n"+settingsMenu.GraphType+"\n"+img.profileURL+"\n"+top5Games), 0744)
	return top5Games, nil
}

//Creates the image
func (img *imageGenerate) createImage() (string, error) {
	dir, _ := os.Getwd()
	command := exec.Cmd{
		Path: path.Join(dir, "genImage", "imageGen"),
		Env: []string{
			"STATIC_DIR=" + img.staticDir,
			"IMAGE_DIR=" + img.dir,
		},
	}
	err := command.Run()
	if err != nil {
		return "", err
	}
	return path.Join(img.dir, "out.png"), nil
}

//Cleans up the dir
func (img *imageGenerate) cleanup() {
	os.RemoveAll(img.dir)
}

//imgClient bounds every image download so a hanging host cannot stall stats
//generation.
var imgClient = &http.Client{Timeout: time.Second * 10}

//download fetches a url and returns its bytes; callers hash, validate and
//save the returned payload, so an icon is never fetched twice.
func download(URL string) ([]byte, error) {
	res, err := imgClient.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("url %s answered status %d", URL, res.StatusCode)
	}
	return ioutil.ReadAll(res.Body)
}

//testImg checks that a downloaded payload decodes as an image.
func testImg(data []byte) bool {
	_, _, err := image.Decode(bytes.NewReader(data))
	return err == nil
}

//testImgURL is the production probe for getImage: the url must download and
//the payload must decode.
func testImgURL(URL string) bool {
	data, err := download(URL)
	return err == nil && testImg(data)
}

//hashBytes returns the base64 url safe sha1 of data, used to detect the
//image behind a cached url having changed.
func hashBytes(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

//IGDB returns protocol relative cover urls (//images.igdb.com/...), make them fetchable
func normalizeURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return rawURL
}

//Gets the image from the IGDB search results. The valid probe checks that a
//candidate url actually serves a decodable image; it is injected so tests
//need no network.
func getImage(games []igdbGame, gameName string, valid func(URL string) bool) string {
	//Prefers the cover of the game whose name matches exactly, falling back to
	//the first cover that passes the probe.
	fallback := ""
	for i := range games {
		imgURL := normalizeURL(games[i].Cover.URL)
		if imgURL == "" {
			continue
		}
		if !valid(imgURL) {
			continue
		}
		if strings.EqualFold(games[i].Name, gameName) {
			return imgURL
		}
		if fallback == "" {
			fallback = imgURL
		}
	}
	if fallback != "" {
		return fallback
	}
	//Nothing matched; the caller substitutes the bundled placeholder icon.
	return ""
}
