#!/usr/bin/env bash

rsync -a --rsh "ssh $1@users.isi.deterlab.net ssh" $2 $3
