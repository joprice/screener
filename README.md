# screener
service serving webpage screenshots

##usage

```
Usage of screener:
  -mongo-host="127.0.0.1": mongo host
  -port=8080: http port
  -web-driver-url="http://127.0.0.1:4444": web driver url
```

##running
* run a selenium server, e.g. `phantomjs --webdriver=4444 --web-security=false --ssl-protocol=any --ignore-ssl-errors=true`
* run `screener`
* hit it with a url `http://localhost:8080/screenshot?url=http%3A%2F%2Fwww.google.com`

