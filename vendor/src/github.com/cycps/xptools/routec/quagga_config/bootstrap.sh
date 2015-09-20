#/usr/bin/env bash

sudo apt-get update
sudo apt-get install -y golang tmux quagga

cd /tmp/cypress
chmod +x do-quagga-config.sh
./do-quagga-config.sh
