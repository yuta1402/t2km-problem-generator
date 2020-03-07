FROM golang:stretch

ENV PHANTOMJS_VERSION 2.1.1-linux-x86_64

RUN apt update && \
    apt install -y --no-install-recommends \
        ca-certificates \
        bzip2 \
        libfontconfig \
        curl \
        make && \
    apt clean

RUN curl -L https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-${PHANTOMJS_VERSION}.tar.bz2 | \
    tar xj -C /usr/local/bin phantomjs-${PHANTOMJS_VERSION}/bin/phantomjs --strip=2

CMD ["/bin/sh"]
