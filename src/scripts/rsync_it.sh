#!/usr/bin/env bash

rsync -a -e "ssh -o StrictHostKeyChecking=no" $2 $1@$3
