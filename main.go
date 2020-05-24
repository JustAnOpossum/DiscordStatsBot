//Main go file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/api"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var bot *session.Session
var totalGuilds int
var guildBlacklist []string

var statsCollection *mongo.Collection
var iconCollection *mongo.Collection
var settingCollection *mongo.Collection

//The steps that main takes to start are
//1. Connects to discord and set's up bot.
//2. Adds the handlers for discord
func main() {
	//Sets up the connection to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	//Pings the server to make sure it is online
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	//Adds the collections to the global vars so the program can access them
	statsCollection = client.Database("statbot").Collection("gamestats")
	iconCollection = client.Database("statbot").Collection("gameicons")
	settingCollection = client.Database("statbot").Collection("settings")

	//Sets up a bool so that indexes can be created with no duplicates
	noRepeats := true
	//Sets the index for the stats collection
	statsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1, "game": 1}, Options: &options.IndexOptions{Unique: &noRepeats}})
	//Sets the index for the icon collection
	iconCollection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"name": 1}, Options: &options.IndexOptions{Unique: &noRepeats}})
	//Sets the index for the settings collection
	settingCollection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: &options.IndexOptions{Unique: &noRepeats}})

	//Set's up the connection to discords api.
	session, err := session.New("Bot " + os.Getenv("BOT_TOKEN"))
	//Identifies intents to receive
	//1 is guilds
	//256 is guild precences
	//512 is guild messages
	//4096 is dms
	session.Gateway.Identifier.Intents = 256 + 1 + 512 + 4096
	bot = session
	if err != nil {
		panic(err)
	}
	err = session.Open()
	if err != nil {
		panic(err)
	}

	defer session.Close()
	session.Gateway.UpdateStatus(gateway.UpdateStatusData{
		Game: &discord.Activity{
			Name: "@ to get stats",
		},
	})
	//Switched between normal status and status displaying tracked servers
	statusUpdate := time.NewTicker(time.Second * 10)
	flip := false
	go func() {
		for {
			select {
			case <-statusUpdate.C:
				var playingStr string
				if flip {
					// playingStr = "Tracking stats for " + strconv.Itoa(totalGuilds) + " servers!"
					playingStr = "NEW UPDATE! Custom status is working + new settings"
					flip = false
				} else {
					playingStr = "@ to get stats"
					flip = true
				}
				session.Gateway.UpdateStatus(gateway.UpdateStatusData{
					Game: &discord.Activity{
						Name: playingStr,
					},
				})
			}
		}
	}()

	//Adds handalers for bot
	session.AddHandler(presenceUpdate)
	session.AddHandler(guildAdded)
	session.AddHandler(newMessage)

	//Loads guild blacklist
	blacklists := os.Getenv("BLACKLIST")
	guildBlacklist = strings.Split(blacklists, ",")

	fmt.Println("Bot is started :D")

	//Waits for the program to get a signal to close
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-exitChan
}

func newMessage(m *gateway.MessageCreateEvent) {
	if m.Author.Bot {
		return
	}
	for i := range guildBlacklist {
		if guildBlacklist[i] == m.GuildID.String() {
			return
		}
	}

	if m.GuildID == 0 {
		//Two cases, if settings in in progress than hand it to the settings menu and if it isn't send initial message and start menu
		if _, ok := menus[m.Author.ID.String()]; ok {
			if menus[m.Author.ID.String()].settingChange == "" {
				msg := menus[m.Author.ID.String()].handleMsgInit(m)
				bot.SendMessage(m.ChannelID, msg, nil)
			} else {
				msg := menus[m.Author.ID.String()].handleMsgSetting(m)
				bot.SendMessage(m.ChannelID, msg, nil)
			}
		} else {
			msg := startMenu(m)
			bot.SendMessageComplex(m.ChannelID, api.SendMessageData{Embed: msg})
		}
	} else {
		//Inital checks to make sure the bot is mentioned
		if len(m.Mentions) == 0 {
			return
		}
		if me, _ := bot.Me(); me.ID != m.Mentions[0].ID {
			return
		}

		//Sends a welcome message to know that the stats are being generated
		welcomeMsg, _ := bot.SendMessage(m.ChannelID, "Creating Your Stats, Please Wait...", nil)

		//Var to see what user was mentioned
		var mentionedUser discord.Snowflake

		//User getting their own stats
		if len(m.Mentions) == 1 {
			mentionedUser = m.Author.ID
		}
		//User getting stats for another user
		if len(m.Mentions) == 2 {
			mentionedUser = m.Mentions[1].User.ID
		}

		//Gets the member from the snowflake
		member, err := bot.Member(m.GuildID, mentionedUser)
		currentDir, _ := os.Getwd()

		image := imageGenerate{
			dir:        path.Join(os.TempDir(), m.Author.ID.String()),
			staticDir:  path.Join(currentDir, "genImage", "Static"),
			profileURL: member.User.AvatarURL(),
			userID:     member.User.ID.String(),
			name:       member.User.Username,
		}
		//Sets up the image to be created
		imgCaption := "Here your stats!"
		var imagePath string
		top5, err := image.setup()
		top5 = strings.ReplaceAll(top5, "\n", " ")
		imgCaption += " " + top5
		if err != nil {
			imagePath = path.Join(currentDir, "genImage", "Static", "avatarError.png")
			imgCaption = "An error occured in image setup: " + err.Error() + "\nPlease report this error to NerdyRedPanda#7480"
		}
		imagePath, err = image.createImage()
		if err != nil {
			imagePath = path.Join(currentDir, "genImage", "Static", "avatarError.png")
			imgCaption = "An error occured in image gen: " + err.Error() + "\nPlease report this error to NerdyRedPanda#7480"
		}
		file, err := os.Open(imagePath)
		defer file.Close()
		bot.SendMessageComplex(m.ChannelID, api.SendMessageData{
			Content: imgCaption,
			Files: []api.SendMessageFile{
				{
					Name:   "Stats.png",
					Reader: file,
				},
			},
		})
		// image.cleanup()
		bot.DeleteMessage(m.ChannelID, welcomeMsg.ID)
	}
}

//Called when guild is created, used to track how many guilds the bot is in
func guildAdded(g *gateway.GuildCreateEvent) {
	totalGuilds++
}

//Handles presence update
func presenceUpdate(p *gateway.PresenceUpdateEvent) {
	//Inital checks to weed out bad data
	activities := p.Activities
	userID := p.User.ID.String()
	if p.User.Bot {
		return
	}
	for i := range guildBlacklist {
		if guildBlacklist[i] == p.GuildID.String() {
			return
		}
	}

	//Adds user to the container if does not exist
	if users.exists(userID) != true {
		users.add(userID)
		users.get(userID).createSettings()
	}

	//Goes through the activities to find new games to start tracking
	for i := range activities {
		if activities[i].Name != "Custom Status" {
			if !users.get(userID).gameExists(activities[i].Name) {
				users.get(userID).startPlaying(activities[i].Name)
			}
		}
	}

	//Checks to see if games are missing and therefore stopped playing
	for name := range users.get(userID).currentGames {
		var found bool
		for i := range activities {
			if activities[i].Name == name {
				found = true
			}
		}
		//Only called if the game is not found in the activities
		if !found {
			users.get(userID).stopPlaying(name)
		}
	}
}
