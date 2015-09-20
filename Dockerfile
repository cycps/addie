FROM ubuntu:15.04
RUN apt-get update && apt-get install -y \
  vim \
  tmux \
  bind9utils \
  libc++1 \
  libc++abi1 \
  uuid-runtime \
  rsync \
  ssh

ADD bin/addie /usr/bin/addie
ADD bin/gatekeeper /usr/bin/gatekeeper

RUN mkdir -p /cypress/keys
ADD devkeys/ssl/* /cypress/keys/

RUN mkdir -p /root/.ssh
COPY devkeys/ssh/* /root/.ssh/

RUN mkdir -p /cypress/scripts
ADD src/scripts/* /cypress/scripts/

RUN bash -c "echo '206.117.25.50 spi.deterlab.net' >> /etc/hosts"

CMD gatekeeper
