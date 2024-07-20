#!/bin/bash

cd $(dirname "$0") || exit

docker stop telegram-bot || true
docker rm telegram-bot || true

docker pull khodand/telegram-bot:latest

cp bot.log "logs/telegram-bot_$(date +%F).log"

nohup docker run -v ./config:/config --name tgbot khodand/telegram-bot:latest 1>bot.log 2>&1 &
