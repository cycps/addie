#!/usr/bin/env bash

dredd ~/.cypress/addie/spec/design.md https://192.168.1.202:8080 --hookfiles=$HOME/.cypress/addie/spec/hooks*.js
