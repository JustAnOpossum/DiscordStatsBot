//Methods for creating the image and downloading the game images

package main

import (
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"net/http"
)

type imageGenerate struct {
	dir        string
	profileURL string
}

//Sets up the dir for image generation
func (img *imageGenerate) setup() {

}

//Creates the image
func (img *imageGenerate) createImage() (string, error) {
	return "", nil
}

//Cleans up the dir
func (img *imageGenerate) cleanup() {

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
