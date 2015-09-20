FROM ubuntu:15.04
RUN apt-get update && apt-get install -y \
  vim \
  tmux \
  bind9utils \
  libc++1 \
  libc++abi1 \
  uuid-runtime

ADD bin/addie /usr/bin/addie
ADD bin/gatekeeper /usr/bin/gatekeeper

RUN mkdir -p /cypress/keys
ADD devkeys/* /cypress/keys/

RUN bash -c "echo '206.117.25.50 spi.deterlab.net' >> /etc/hosts"

CMD gatekeeper
