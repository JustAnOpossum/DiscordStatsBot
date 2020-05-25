package main

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestStartPlaying(t *testing.T) {
	users.add("123")
	users.get("123").startPlaying("game1")
	if _, ok := users.get("123").currentGames["game1"]; !ok {
		t.Error("Game is not game1")
	}
	users = container{users: make(map[string]*user)}
}

func TestStopPlaying(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DB_STRING")))
	if err != nil {
		t.Error("Error connecting to database")
	}
	defer client.Disconnect(ctx)
	//Pings the server to make sure it is online
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		t.Error("Error pinging server")
	}
	//Adds the collections to the global vars so the program can access them
	statsCollection = client.Database("testing1").Collection("gamestats")
	iconCollection = client.Database("testing1").Collection("gameicons")
	settingCollection = client.Database("testing1").Collection("settings")
	defer statsCollection.Drop(ctx)
	defer iconCollection.Drop(ctx)
	defer settingCollection.Drop(ctx)

	users.add("123")
	users.get("123").startPlaying("game1")
	time.Sleep(time.Second * 3)
	users.get("123").stopPlaying("game1")
	var originalStat stat
	err = statsCollection.FindOne(ctx, bson.M{"id": "123"}).Decode(&originalStat)
	if err != nil {
		t.Error("Error looing up item " + err.Error())
	}
	if originalStat.Hours == 0 {
		t.Error("Time was 0")
	}

	users.get("123").startPlaying("game1")
	time.Sleep(time.Second * 3)
	users.get("123").stopPlaying("game1")

	var secondStat stat
	err = statsCollection.FindOne(ctx, bson.M{"id": "123"}).Decode(&secondStat)
	if err != nil {
		t.Error("Error looing up item")
	}
	if secondStat.Hours <= originalStat.Hours {
		t.Error("Second stat is = or < than new one")
	}
}
