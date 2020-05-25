package main

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestGetHash(t *testing.T) {
	hash, _ := getHash("https://red-panda.me/img/botError.png")
	if hash != "h9FxvZ2mhQucUimxYcohhf3s79U=" {
		t.Error("Hash was not correct, hash was " + hash)
	}
}

func TestImageDecodingCorrent(t *testing.T) {
	result := testImg("https://red-panda.me/img/botError.png")
	if !result {
		t.Error("testImg failed when correct image was inputted")
	}
}

func TestImageDecodingIncorrent(t *testing.T) {
	result := testImg("https://red-panda.me/")
	if result {
		t.Error("testImg passed when it should of failed")
	}
}

func TestGetImgWithImg(t *testing.T) {
	mockBing := bingAnswer{
		Value: []bingValue{
			{EncodingFormat: "png", ContentURL: "https://red-panda.me/img/connor10.9.png"},
		},
	}
	URL := getImage(mockBing)
	if URL != "https://red-panda.me/img/connor10.9.png" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}

func TestGetImgWithNo(t *testing.T) {
	mockBing := bingAnswer{}
	URL := getImage(mockBing)
	if URL != "https://red-panda.me/img/botError.png" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}

func TestImageSetup(t *testing.T) {
	currentDir, _ := os.Getwd()
	image := imageGenerate{
		dir:        path.Join(os.TempDir(), "123"),
		staticDir:  path.Join(currentDir, "genImage", "Static"),
		profileURL: "https://red-panda.me/img/full/3g5wej.png",
		userID:     "123",
		name:       "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Error("Error connecting to database " + os.Getenv("test") + " after")
	}
	defer client.Disconnect(ctx)
	//Pings the server to make sure it is online
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		t.Error("Error pinging server")
	}
	//Adds the collections to the global vars so the program can access them
	statsCollection = client.Database("testing").Collection("gamestats")
	iconCollection = client.Database("testing").Collection("gameicons")
	settingCollection = client.Database("testing").Collection("settings")
	defer statsCollection.Drop(ctx)
	defer iconCollection.Drop(ctx)
	defer settingCollection.Drop(ctx)

	_, err = statsCollection.InsertOne(ctx, stat{ID: "123", Game: "game1", Hours: 1, Ignore: false})
	if err != nil {
		t.Error("Error inserting stat " + err.Error())
		return
	}

	_, err = iconCollection.InsertOne(ctx, icon{Name: "game1", URL: "https://images-wixmp-ed30a86b8c4ca887773594c2.wixmp.com/intermediary/f/571e5943-4616-4654-bf99-10b3c98f8686/d98301o-426f05ca-8fe5-4636-9009-db9dd1fca1f3.png", Hash: "I0RkhNbblYgC3xUn0t3IJmPKNfI="})
	if err != nil {
		t.Error("Error inserting icon " + err.Error())
		return
	}

	_, err = image.setup()
	if err != nil {
		t.Error("Error creating image folder " + err.Error())
	}
	defer image.cleanup()

	if _, err := os.Stat(image.dir); err != nil {
		t.Error("Error reading dir " + err.Error())
	}

	if _, err := os.Stat(path.Join(image.dir, "data")); err != nil {
		t.Error("Error reading data " + err.Error())
	}
	if _, err := os.Stat(path.Join(image.dir, "game1")); err != nil {
		t.Error("Error reading game1 " + err.Error())
	}
	if _, err := os.Stat(path.Join(image.dir, "game1", "hours")); err != nil {
		t.Error("Error reading game1 hours " + err.Error())
	}
	if _, err := os.Stat(path.Join(image.dir, "game1", "icon")); err != nil {
		t.Error("Error reading game1 icon " + err.Error())
	}
}
