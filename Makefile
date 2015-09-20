.PHONY: all
all: services container

services = bin/addie, bin/gatekeeper

services:
	gb build

.PHONY: container
container:
	docker build -t addie .

.PHONY: run
run:
	docker run -d -p 8081:8081 --hostname=addie --name=addie --link=data --add-host=spi.deterlab.net:206.117.25.50 addie

.PHONY: debug
debug:
	docker run -i -t -p 8081:8081 --hostname=addie --name=addie --link=data --add-host=spi.deterlab.net:206.117.25.50 addie || echo "\n"
