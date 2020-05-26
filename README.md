# [DiscordStatsBot](https://discordapp.com/oauth2/authorize?client_id=461294052529143825&scope=bot&permissions=0) [![Build Status](https://travis-ci.com/NerdyRedPanda/DiscordStatsBot.svg?branch=master)](https://travis-ci.com/NerdyRedPanda/DiscordStatsBot)

<img src="https://red-panda.me/img/botIcon.png" alt="drawing" width="200"/>

## A bot for discord that tracks your games and displays them in a graph

<img src="https://red-panda.me/img/statBot/Stats.png" alt="drawing" width="500"/>

# **Usage**

### You just have to mention the bot to get your own stats.

<img src="https://red-panda.me/img/statBot/gettingStats.gif" alt="drawing" />

You can get stats for another member by mentioning the bot and then mentioning them.

# Settings Menu

To access the settings menu all you have to do is send a DM to the bot and it will let you change your settings.

<img src="https://red-panda.me/img/statBot/Settings.png" alt="drawing" />

## graph (bar, pie): Your graph type for the images.
## hide: Allows you to hide games from your graph.
## show: Allows you to show your hidden games
## mention (enabled, disabled): Allows someone else to get your stats by mentioning you.
## delete: Deletes all the data the bot has for you.

# Bot Info

## Environment Variables:

BOT_TOKEN: Token for the bot

BING_KEY: Key for the bing image search

BLACKLIST: Comma seperated list of guild IDs to ignore updates from

DB_STRING: Connection string for mongodb

## Build info:

The go program in genImage must be built. It requires imagemagick 7 compiled from source to be built corrently.
