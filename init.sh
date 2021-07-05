#!/bin/bash

rm vkBot

git pull origin master

 go build -o vkBot -a

./vkBot
