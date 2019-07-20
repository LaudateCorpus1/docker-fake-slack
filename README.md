# Fake Slack

A minimal docker image to fake the Slack API. Currently only concerned with the very minimum required to test Canvas LMS locally.

## Getting started

```
$ docker run -it -p 9393:9393 -e VIRTUAL_HOST=slack.docker ahuff44/fake-slack
```

Note that the `VIRTUAL_HOST` environment variable will let you access the service
at the host `slack.docker`, if you are using `dinghy` as part of your docker setup.

The server will log all requests it recieves to disk. The files will be placed at
`/messages/slack/{{route}}_{{ts}}`; e.g. `/messages/slack/chat.postMessage_1563577444.405595`

To use with docker-compose, here is a basic service block to start from:
```yaml
  slack:
    image: ahuff44/fake-slack
    volumes:
      - "/messages/slack"
    ports:
      - "9393"
```

## API

The api is as simple as possible and returns dummy data.
Some endpoints (especially `users.lookupByEmail`) don't include a full
response. See [the official documentation](https://api.slack.com/methods/users.lookupByEmail)
for descriptions of what these endpoints actually return in the real slack API.

`auth.test`:
```bash
$ curl slack.docker/api/auth.test -d'token=foo'
{"ok":true}
$ curl slack.docker/api/auth.test -d'token='
{"ok":false,"error":"not_authed"}
```

`users.lookupByEmail`:
```bash
$ curl slack.docker/api/users.lookupByEmail -d'token=foo&email=bob@example.com'
{"ok":true,"user":{"id":"UXXXXXXXX"}}
$ curl slack.docker/api/users.lookupByEmail -d'token=foo&email=bademail'
{"ok":false,"error":"users_not_found"}
$ curl slack.docker/api/users.lookupByEmail -d'token=foo'
{"ok":false,"error":"users_not_found"}
```

`im.open`:
```bash
$ curl slack.docker/api/im.open -d'token=foo&user=U12345678'
{"ok":true,"channel":{"id":"DXXXXXXXX"}}
$ curl slack.docker/api/im.open -d'token=foo&user=Ubad'
{"ok":false,"error":"user_not_found"}
$ curl slack.docker/api/im.open -d'token=foo&user=verybad'
{"ok":false,"error":"user_not_found"}
$ curl slack.docker/api/im.open -d'token=foo'
{"ok":false,"error":"user_not_found"}
```

`chat.postMessage`:
```bash
$ curl slack.docker/api/chat.postMessage -d'token=foo&text=hi&channel=bar'
{"ok":true,"channel":"bar","ts":"1563567103.077513","message":{"type":"message","subtype":"bot_message","text":"hi","ts":"1563567103.077513","username":"default-username","bot_id":"BXXXXXXXX"}}
$ curl slack.docker/api/chat.postMessage -d'token=foo&channel=DXXXXXXXX&icon_url=http%3A%2F%2Florempixel.com%2F48%2F48%2F&text=This+is+a+test+message&username=canvas&attachments=%5B%7B%22fallback%22%3A%22%22%2C%22text%22%3A%22And+this+is+an+attachment%21%22%7D%2C%7B%22fallback%22%3A%22%22%2C%22text%22%3A%22another%21%22%7D%5D'
{"ok":true,"channel":"DXXXXXXXX","ts":"1563567115.463191","message":{"type":"message","subtype":"bot_message","text":"This is a test message","ts":"1563567115.463191","username":"canvas","bot_id":"BXXXXXXXX","attachments":"[{\"fallback\":\"\",\"text\":\"And this is an attachment!\"},{\"fallback\":\"\",\"text\":\"another!\"}]"}}
$ curl slack.docker/api/chat.postMessage -d'token=foo&text=hi'
{"ok":false,"error":"channel_not_found"}
$ curl slack.docker/api/chat.postMessage -d'token=foo&channel=bar'
{"ok":false,"error":"no_text"}
```
