# pullanusbot

[![Build Status](https://github.com/ailinykh/pullanusbot/workflows/build/badge.svg)](https://github.com/ailinykh/pullanusbot/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/ailinykh/pullanusbot)](https://goreportcard.com/report/github.com/ailinykh/pullanusbot)
![GitHub](https://img.shields.io/github/license/ailinykh/pullanusbot.svg)

This bot helps your telegram chat to consume content in more native way

Let's say somebody sends a link to the webm video:

![bot-webm-video](https://user-images.githubusercontent.com/939390/95298451-c7757100-0884-11eb-9140-4c6474959720.gif)

Or a video file as a document:

![bot-video-convert](https://user-images.githubusercontent.com/939390/95298623-07d4ef00-0885-11eb-92e4-b3c2015f7ecc.gif)

It's even support links to twitter videos

![bot-twitter-video](https://user-images.githubusercontent.com/939390/95298730-3783f700-0885-11eb-9650-b0c04e40aa2f.gif)

... and images!

![bot-twitter-images](https://user-images.githubusercontent.com/939390/95298790-4cf92100-0885-11eb-8bb2-8adbc91f5b23.gif)

## how to run

Setup environment

```shell
brew install go ffmpeg youtube-dl
```
clone repo

```shell
git clone https://github.com/ailinykh/pullanusbot.git
cd pullanusbot
```

install go dependencies
```shell
go mod download
```
obtain bot token from [@BotFather](https://t.me/BotFather) and your telegram ID via `/info` command from [@pullanusbot](https://t.me/pullanusbot)

```shell
echo "export BOT_TOKEN:=12345678:XXXXXXXXxxxxxxxxXXXXXXXXxxxxxxxxXXX" > .env
echo "export ADMIN_CHAT_ID:=123456789" >> .env
```

and run!

```shell
make
```
