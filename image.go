//Methods for creating the image and downloading the game images
//Sets up the dir for images and creates the images, and removes the image
//Also has methods for downloading the images and finding the images from a bing search result.

package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
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

//Gets and saves an image, function so it can be called more than once
func getAndSaveImg(ctx context.Context, gameName string, imgDir string) error {
	searchResults, err := search(gameName)
	if err != nil {
		return err
	}
	imgURL := getImage(searchResults)
	hash, err := getHash(imgURL)
	if err != nil {
		return err
	}
	_, err = iconCollection.InsertOne(ctx, bson.M{"name": gameName, "url": imgURL, "hash": hash})
	res, err := http.Get(imgURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	ioutil.WriteFile(path.Join(imgDir, gameName, "icon"), body, 0744)
	return nil
}

//Sets up the dir for image generation
func (img *imageGenerate) setup() (string, error) {
	err := os.Mkdir(img.dir, 0744)

	ctx, close := context.WithTimeout(context.Background(), time.Second*15)
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
		err = os.Mkdir(path.Join(img.dir, stats[i].Game), 0744)
		//Writes the hours
		ioutil.WriteFile(path.Join(img.dir, stats[i].Game, "hours"), []byte(strconv.FormatFloat(stats[i].Hours, 'f', -1, 64)), 0744)
		err := iconCollection.FindOne(ctx, bson.M{"name": stats[i].Game}).Decode(&userGame)
		//Case if the game img does not exist
		if err == mongo.ErrNoDocuments {
			err = getAndSaveImg(ctx, stats[i].Game, img.dir)
			if err != nil {
				return "", err
			}
			//Case if it does exist
		} else {
			hash, _ := getHash(userGame.URL)
			//If the hashes for the image don't match
			if hash != userGame.Hash {
				iconCollection.DeleteOne(ctx, bson.M{"name": userGame.Name})
				err = getAndSaveImg(ctx, stats[i].Game, img.dir)
				if err != nil {
					return "", err
				}
			}
			res, err := http.Get(userGame.URL)
			if err != nil {
				return "", err
			}
			defer res.Body.Close()
			body, _ := ioutil.ReadAll(res.Body)
			ioutil.WriteFile(path.Join(img.dir, stats[i].Game, "icon"), body, 0744)
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
	err = settingCollection.FindOne(ctx, bson.M{"id": img.userID}).Decode(&settingsMenu)

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

//Tests to make sure that img is valid
func testImg(URL string) bool {
	res, err := http.Get(URL)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	_, _, err = image.Decode(res.Body)
	if err != nil {
		return false
	}
	return true
}

//Get's the image from the bing search results
func getImage(images bingAnswer) string {
	//First loop checks to see if there is a png image in the array, if there is return it.
	//If not find jpegs and return that.
	for i := range images.Value {
		if images.Value[i].EncodingFormat == "png" {
			if !testImg(images.Value[i].ContentURL) {
				continue
			}
			return images.Value[i].ContentURL
		}
	}
	for i := range images.Value {
		if images.Value[i].EncodingFormat == "jpeg" {
			if !testImg(images.Value[i].ContentURL) {
				continue
			}
			return images.Value[i].ContentURL
		}
	}
	return "https://red-panda.me/img/botError.png"
}

//Gets the hash of the image
func getHash(URL string) (string, error) {
	res, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	hash := sha1.New()
	hash.Write(body)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}
