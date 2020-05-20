//Methods for creating the image and downloading the game images

package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

type imageGenerate struct {
	dir        string
	staticDir  string
	profileURL string
	userID     string
}

//Sets up the dir for image generation
func (img *imageGenerate) setup() error {
	err := os.Mkdir(img.dir, 0744)
	//TODO database call to get data
	err = os.Mkdir(path.Join(img.dir, "Spotify"), 0744)
	if err != nil {
		return err
	}
	ioutil.WriteFile(path.Join(img.dir, "data"), []byte("10\n20\nNerdyRedPanda\nbar\nhttps://red-panda.me/img/full/3g5wej.png\nSpotify"), 0744)
	ioutil.WriteFile(path.Join(img.dir, "Spotify", "hours"), []byte("15"), 0744)

	file, err := os.Create(path.Join(img.dir, "Spotify", "icon"))
	res, err := http.Get("https://red-panda.me/img/full/3g5wej.png")
	if err != nil {

	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	file.Write(body)
	return nil
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
		fmt.Println(err)
	}
	return path.Join(img.dir, "out.png"), nil
}

//Cleans up the dir
func (img *imageGenerate) cleanup() {
	os.RemoveAll(img.dir)
}

//Get's the image from the bing search results
func getImage(gameName string) (string, error) {
	images, err := search(gameName)
	if err != nil {
		return "", err
	}
	//First loop checks to see if there is a png image in the array, if there is return it.
	//If not find jpegs and return that.
	for i := range images.Value {
		if images.Value[i].EncodingFormat == "png" {
			return images.Value[i].ContentURL, nil
		}
	}
	for i := range images.Value {
		if images.Value[i].EncodingFormat == "jpeg" {
			return images.Value[i].ContentURL, nil
		}
	}
	return "error", nil
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
