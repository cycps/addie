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
	docker run -d -p 8081:8081 --hostname=addie --name=addie --add-host=spi.deterlab.net:206.117.25.50 addie

.PHONY: debug
debug:
	docker run -i -t \
		-p 8081:8081 \
		-v `cd ..; pwd`:/builder \
		--hostname=addie --name=addie \
		--entrypoint bash \
		--add-host=spi.deterlab.net:206.117.25.50 addie || echo "\n"
