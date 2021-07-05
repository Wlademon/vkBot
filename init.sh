#!/bin/bash

sleep 1000

rm vkBot

git pull origin master

go build -o vkBot -a

./vkBot
