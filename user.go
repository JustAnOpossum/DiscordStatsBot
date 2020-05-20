//User struct and methods.
//User is the core component since everyone with the bot is a "user"

package main

import (
	"fmt"
	"sync"
	"time"
)

//Global to hold the container so it can be accessed by any method
var users = container{users: make(map[string]*user)}

type game struct {
	name        string
	timeStarted time.Time
}

type user struct {
	userID       string
	currentGames map[string]game
	mutex        sync.Mutex
}

//Method called when a user starts playing a game
func (u *user) startPlaying(name string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if u.gameExists(name) {
		return
	}
	u.currentGames[name] = game{name: name, timeStarted: time.Now()}
}

//Method called when a user stops playing the game
func (u *user) stopPlaying(name string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if !u.gameExists(name) {
		return
	}
	time := time.Now().Sub(u.currentGames[name].timeStarted)
	u.saveTime(time.Hours())
	delete(u.currentGames, name)
}

//Saves the time to the database
func (u *user) saveTime(time float64) {
	fmt.Println(time)
}

//Checks to see if a game exists in their currently playing games
func (u *user) gameExists(name string) bool {
	if _, ok := u.currentGames[name]; ok {
		return true
	}
	return false
}

//Struct to hold the users and methods to add and get
type container struct {
	mutex sync.Mutex
	users map[string]*user
}

//Creates a new user and adds it to the container
func (c *container) add(userID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.exists(userID) == true {
		return
	}
	c.users[userID] = &user{userID: userID, currentGames: make(map[string]game)}
}

//Checks to see if the user exists
func (c *container) exists(userID string) bool {
	if _, ok := c.users[userID]; ok {
		return true
	}
	return false
}

func (c *container) get(userID string) *user {
	return c.users[userID]
}
