#!/bin/bash

if [ $1 -eq "refs/refs/heads/master"]; do
    git pull origin
fi