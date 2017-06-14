## Fill out manifest and push the app using cf push

```
---
applications:
  - name: dnsprober
    memory: 128M
    instances: 1
    buildpack: https://github.com/kr/heroku-buildpack-go.git
    command: dnsprober
    env:
      PINGHOST: "some.host.domain"
      INTERVAL: 15
      DNS_SERVERS: "1.1.1.1,2.2.2.2"
```
