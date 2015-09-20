#!/usr/bin/env bash

cp head /etc/resolvconf/resolv.conf.d/
resolvconf -u
