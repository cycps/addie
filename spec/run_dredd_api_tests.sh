#!/usr/bin/env bash

dredd ~/apitest/design.md https://192.168.1.202:8080 --hookfiles=./hooks*.js
