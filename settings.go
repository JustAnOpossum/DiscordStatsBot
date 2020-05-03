//settings.go
//Contains methods for settings menu
//

package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

//Global to keep track of state of settings menu
var menus = make(map[string]*settingsMenu)

//All available settings to check to see if a message is valid
var settings = [5]string{"graph", "hide", "show", "mention", "delete"}

//Main menu message embed object
var mainMenu = &discord.Embed{
	Title: "Settings for Game Stats Bot",
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
func genSetting(setting string) (string, []string) {
	settingStr := setting + " Settings (Type \"cancel\" to cancel)\n\n"
	var options []string
	switch setting {
	case "graph":
		settingStr += "1. Bar\n2. Pie"
		options = []string{"bar", "pie"}
		break
	case "hide":
		//TODO: Database call here to populate string
		break
	case "show":
		//TODO: Database call here to populate string
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
		settingStr, sliceOptions := genSetting(m.Content)
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
	//Case for true false settings
	if _, err := strconv.ParseBool(s.options[option]); err != nil {
		//TODO: Database call to save setting
	}

}
