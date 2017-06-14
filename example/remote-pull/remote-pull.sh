#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Don't run this shell script directly!"
    exit 1
fi

if [ $1 == "refs/heads/master" ]; then
    git pull origin
else
    exit 1
fi