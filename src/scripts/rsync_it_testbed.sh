#!/usr/bin/env bash

rsync -a -e "ssh -o StrictHostKeyChecking=no" --rsh "ssh $1@users.isi.deterlab.net ssh" $2 $3
