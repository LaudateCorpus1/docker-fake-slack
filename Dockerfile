FROM golang:1.12

WORKDIR /usr/src/app/

RUN useradd -u 1000 -M docker \
  && mkdir -p /messages/slack \
  && chown docker /messages/slack
USER docker

VOLUME /messages/slack
EXPOSE 9393

CMD "echo hi"
