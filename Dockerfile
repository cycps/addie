FROM ubuntu:15.04
RUN apt-get update && apt-get install -y \
  vim \
  tmux \
  bind9utils \
  libc++1 \
  libc++abi1 \
  uuid-runtime \
  rsync \
  ssh \
  curl \
  language-pack-en

ADD bin/addie /usr/bin/addie
ADD bin/gatekeeper /usr/bin/gatekeeper

RUN mkdir -p /cypress/keys
ADD devkeys/ssl/* /cypress/keys/

RUN mkdir -p /root/.ssh
COPY devkeys/ssh/* /root/.ssh/

RUN mkdir -p /cypress/scripts
ADD src/scripts/* /cypress/scripts/

RUN bash -c "echo '206.117.25.50 spi.deterlab.net' >> /etc/hosts"
ENV LD_LIBRARY_PATH /usr/local/lib

RUN curl -LO https://github.com/cycps/sim/releases/download/v0.1.alpha.2/CySim-0.1.tar.gz
RUN mkdir sim
RUN mv CySim-0.1.tar.gz sim/
RUN cd sim && tar xzf CySim-0.1.tar.gz && cd CySim-0.1 && ./install.sh

CMD gatekeeper
