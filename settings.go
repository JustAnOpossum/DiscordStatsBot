//settings.go
//Contains methods for settings menu
//starMenu is called first which adds the menu to the global
//handleMsgInit is called to get the inital setting choice and will keep being called unless they enter a valid option
//handleMsgSetting is called when they pick an option
//save is called and then delete is called to remove the menu from the global

package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"go.mongodb.org/mongo-driver/bson"
)

//Global to keep track of state of settings menu
var menus = make(map[string]*settingsMenu)

//All available settings to check to see if a message is valid
var settings = [5]string{"graph", "hide", "show", "mention", "delete"}

//Main menu message embed object
var mainMenu = &discord.Embed{
	Title: "Settings for Game Stats Bot (Contact NerdyRedPanda#7480 with any issues)",
	Fields: []discord.EmbedField{
		{
			Name:   ":bar_chart:",
			Value:  "graph - Changes the graph type",
			Inline: false,
		},
		{
			Name:   ":no_entry_sign:",
			Value:  "hide - Allows you to hide games from your graph",
			Inline: false,
		},
		{
			Name:   ":o:",
			Value:  "show - Allows you to show hidden games",
			Inline: false,
		},
		{
			Name:   ":bell:",
			Value:  "mention - Allows other people to get your stats by mentioning you",
			Inline: false,
		},
		{
			Name:   ":wastebasket:",
			Value:  "delete - Deletes all your data.\nWARNING: This cannot be undone",
			Inline: false,
		},
	},
}

type settingsMenu struct {
	inMenu        bool
	userID        string
	settingChange string
	options       []string
	timer         *time.Timer
}

//Starts the settings menu flow
func startMenu(m *gateway.MessageCreateEvent) *discord.Embed {
	s := &settingsMenu{
		timer:  time.NewTimer(time.Minute * 2),
		inMenu: true,
		userID: m.Author.ID.String(),
	}
	//Called to cleanup when either timer runs out or setting is changed
	s.timer = time.AfterFunc(time.Minute*2, func() {
		delete(menus, s.userID)
	})
	menus[m.Author.ID.String()] = s
	return mainMenu
}

func stopMenu(s *settingsMenu) {
	s.timer.Stop()
	delete(menus, s.userID)
}

//Generates the main menu
func genSetting(setting string, userID string) (string, []string) {
	settingStr := setting + " Settings (Type \"cancel\" to cancel)\n\n"
	var options []string
	switch setting {
	case "graph":
		settingStr += "1. Bar\n2. Pie"
		options = []string{"bar", "pie"}
		break
	case "hide":
		var hiddenGames []stat
		ctx, close := context.WithTimeout(context.Background(), time.Second*5)
		defer close()
		cursor, _ := statsCollection.Find(ctx, bson.M{"id": userID, "ignore": false})
		cursor.All(ctx, &hiddenGames)
		for i := range hiddenGames {
			settingStr += strconv.Itoa(i+1) + ". " + hiddenGames[i].Game
			options = append(options, hiddenGames[i].Game)
		}
		break
	case "show":
		var shownGames []stat
		ctx, close := context.WithTimeout(context.Background(), time.Second*5)
		defer close()
		cursor, _ := statsCollection.Find(ctx, bson.M{"id": userID, "ignore": true})
		cursor.All(ctx, &shownGames)
		for i := range shownGames {
			settingStr += strconv.Itoa(i+1) + ". " + shownGames[i].Game
			options = append(options, shownGames[i].Game)
		}
		break
	case "mention":
		settingStr += "1. Mentions Enabled\n2. Mentions Disabled"
		options = []string{"true", "false"}
		break
	case "delete":
		settingStr += "1. Delete all data, CANNOT BE UNDONE"
		options = []string{"delete"}
		break
	}
	return settingStr, options
}

//Handles the inital message to change settings
func (s *settingsMenu) handleMsgInit(m *gateway.MessageCreateEvent) string {
	var found bool
	m.Content = strings.ToLower(m.Content)
	for i := 0; i < len(settings); i++ {
		if m.Content == settings[i] {
			found = true
		}
	}
	if found {
		s.settingChange = m.Content
		settingStr, sliceOptions := genSetting(m.Content, m.Author.ID.String())
		s.options = sliceOptions
		return settingStr
	}
	return "Error: Invalid setting, please try again\n"
}

//Handles the message when the user is changing a setting
func (s *settingsMenu) handleMsgSetting(m *gateway.MessageCreateEvent) string {
	option, err := strconv.Atoi(m.Content)
	if strings.ToLower(m.Content) == "cancel" {
		stopMenu(s)
		return "Cancelled"
	}
	if err != nil {
		return "Error: Invalid option, please try again"
	}
	if option-1 > len(s.options)-1 {
		return "Error: Invalid option, please try again"
	}

	s.save(option - 1)
	stopMenu(s)
	return "Setting Saved!"
}

//Saves the setting back into the database
func (s *settingsMenu) save(option int) {
	ctx, close := context.WithTimeout(context.Background(), time.Second*5)
	defer close()
	//Switch for changing the setting
	switch s.settingChange {
	case "graph":
		settingCollection.UpdateOne(ctx, bson.M{"id": s.userID}, bson.M{"$set": bson.M{"graphtype": s.options[option]}})
		break
	case "hide":
		statsCollection.UpdateOne(ctx, bson.M{"id": s.userID, "game": s.options[option]}, bson.M{"$set": bson.M{"ignore": true}})
		break
	case "show":
		statsCollection.UpdateOne(ctx, bson.M{"id": s.userID, "game": s.options[option]}, bson.M{"$set": bson.M{"ignore": false}})
		break
	case "mention":
		parsedBool, _ := strconv.ParseBool(s.options[option])
		settingCollection.UpdateOne(ctx, bson.M{"id": s.userID}, bson.M{"$set": bson.M{"mentionforstats": parsedBool}})
		break
	case "delete":
		settingCollection.DeleteOne(ctx, bson.M{"id": s.userID})
		statsCollection.DeleteMany(ctx, bson.M{"id": s.userID})
		delete(users.users, s.userID)
		break
	}
}
