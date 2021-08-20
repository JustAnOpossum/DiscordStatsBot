//Main go file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/session"
	"github.com/diamondburned/arikawa/v2/utils/sendpart"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var bot *session.Session

var totalGuilds int
var guilds map[string]bool
var guildBlacklist []string
var bots map[string]bool

var statsCollection *mongo.Collection
var iconCollection *mongo.Collection
var settingCollection *mongo.Collection

//The steps that main takes to start are
//1. Connects to discord and set's up bot.
//2. Adds the handlers for discord
func main() {
	//Creats the map to store the bots
	bots = make(map[string]bool)
	guilds = make(map[string]bool)

	//Sets up the connection to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DB_STRING")))
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
	//1 is guilds (required to get messages when guilds are created))
	//256 is guild precences (gets presence updates)
	//512 is guild messages (gets mention messages)
	//4096 is dms (gets dm messages)
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
		Activities: []discord.Activity{
			{
				Name: "Hello",
			},
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
					playingStr = "Tracking stats for " + strconv.Itoa(totalGuilds) + " servers!"
					// playingStr = "NEW UPDATE! Custom status is working + new settings"
					flip = false
				} else {
					playingStr = "@ to get stats"
					flip = true
				}
				session.Gateway.UpdateStatus(gateway.UpdateStatusData{
					Activities: []discord.Activity{
						{
							Name: playingStr,
						},
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
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exitChan
}

//Called when a new message comes in
//Handles methods for both the DM channel and the mentioning for stats
func newMessage(m *gateway.MessageCreateEvent) {
	if m.Author.Bot {
		return
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
		var mentionedUser discord.UserID

		//User getting their own stats
		if len(m.Mentions) == 1 {
			mentionedUser = m.Author.ID
		}
		//User getting stats for another user
		if len(m.Mentions) == 2 {
			//Check here to see if the mentioned user has the setting enabled
			var userSettings setting
			ctx, close := context.WithTimeout(context.Background(), time.Second*5)
			defer close()
			settingCollection.FindOne(ctx, bson.M{"id": m.Mentions[1].User.ID.String()}).Decode(&userSettings)
			if !userSettings.MentionForStats {
				return
			}
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
		imgCaption := "Here your stats! \n("
		var imagePath string
		//tops5 is returned so the bot can show the user that their top 5 games are
		top5, err := image.setup()
		top5arr := strings.Split(top5, "\n")
		//Loops through the top5 to seperate them and put them into a usable format
		for i := 0; i < len(top5arr)-1; i++ {
			if i+1 == len(top5arr)-1 {
				imgCaption += top5arr[i] + ")"
			} else {
				imgCaption += top5arr[i] + ", "
			}
		}
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
			Files: []sendpart.File{
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
	//Simple method to check to see if a guild had already been added, in case of API reconnection
	if _, ok := guilds[g.Guild.ID.String()]; !ok {
		totalGuilds++
		guilds[g.Guild.ID.String()] = true
	}

	//Tracks the state for bota to be checked aganst later
	for i := range g.Members {
		if g.Members[i].User.Bot {
			bots[g.Members[i].User.ID.String()] = true
		}
	}
}

//Handles presence update
func presenceUpdate(p *gateway.PresenceUpdateEvent) {
	//Inital checks to weed out bad data
	activities := p.Activities
	userID := p.User.ID.String()
	//The first update for a bot is true so this tracks if a bot is added or becomes online
	if p.User.Bot {
		bots[p.User.ID.String()] = true
	}
	//Checks to see if a user is a known bot
	if _, ok := bots[p.User.ID.String()]; ok {
		return
	}
	//Makes sure the guild it is from is not on the blacklist (For bot list servers)
	for i := range guildBlacklist {
		if guildBlacklist[i] == p.GuildID.String() {
			return
		}
	}
	var checkUser setting
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	settingCollection.FindOne(ctx, bson.M{"id": userID}).Decode(&checkUser)
	if checkUser.Disable {
		return
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
