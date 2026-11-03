# [DiscordStatsBot](https://discordapp.com/oauth2/authorize?client_id=461294052529143825&scope=bot&permissions=0)

<img src="https://greyopossum.net/img/botIcon.png" alt="drawing" width="200"/>

## A bot for discord that tracks your games and displays them in a graph

<img src="https://greyopossum.net/img/statBot/Stats.png" alt="drawing" width="500"/>

# **Usage**

### You just have to mention the bot to get your own stats.

<img src="https://greyopossum.net/img/statBot/gettingStats.gif" alt="drawing" />

You can get stats for another member by mentioning the bot and then mentioning them.

# Settings Menu

To access the settings menu all you have to do is send a DM to the bot and it will let you change your settings.

<img src="https://greyopossum.net/img/statBot/Settings.png" alt="drawing" />

## graph (bar, pie): Your graph type for the images.
## hide: Allows you to hide games from your graph.
## show: Allows you to show your hidden games
## mention (enabled, disabled): Allows someone else to get your stats by mentioning you.
## delete: Deletes all the data the bot has for you.

# Bot Info

## Environment Variables:

APP_ENV: Which environment file to load (dev or prod). Defaults to dev.

APP_ID: Discord application ID for the bot (differs per environment)

BOT_TOKEN: Token for the bot

BING_KEY: Key for the bing image search

BLACKLIST: Comma seperated list of guild IDs to ignore updates from

DB_STRING: Connection string for mongodb

## Build & development

Both Go programs must be built. `genImage/imageGen` cgo-compiles against
**ImageMagick 7**, so the build needs ImageMagick compiled from source. Rather
than install it by hand, build and develop inside containers — a single
`Dockerfile` produces both the small production image and a full dev toolchain
(the old VS Code devcontainer is gone; use the compose file instead).

Production / CI image (this is what CI builds and pushes, and what the
production compose runs):

```bash
podman build -t discordstatsbot .      # default target = runtime
```

Development environment (Go + ImageMagick toolchain + MongoDB, repo bind-mounted at `/app`):

```bash
podman-compose -p dsb-dev -f docker-compose.dev.yml up -d --build
podman-compose -p dsb-dev -f docker-compose.dev.yml exec dev bash
#   go test ./...
```

The ImageMagick version is pinned as `IMAGEMAGICK_VERSION` in the `Dockerfile`.
