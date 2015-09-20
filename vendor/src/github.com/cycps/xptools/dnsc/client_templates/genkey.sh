#/usr/bin/env bash

dnssec-keygen -a HMAC-MD5 -b 512 -n USER $1
