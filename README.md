aqi indicator for for tmux. powered by the [iqair api](https://www.iqair.com/us/air-pollution-data-api)
> you could just `curl -s "http://api.airvisual.com/v2/nearest_city?key=$IQAIR_API_KEY" | jq ".data | { location: .city, aqi: .current.pollution.aqius }"`

### usage
```
$ go get -v -u github.com/cqsd/tmux-aqi
$ mkdir -p ~/.iq-air
$ echo '132cb485...' > ~/.iq-air/key
$ tmux-aqi
#[fg=brightwhite,bg=red] Noe Valley AQI: 99
```

set it (for example) in `status-right`
```
set -g status-right "#(tmux-aqi) %D %H:%M "
```

![screenshot](./screenshot.jpg)

the most recent run is cached for 5 minutes in `~/.iq-air/lastrun.json` to
respect rate limit (and also because aqi doesn't update very often. in fact it
could wait longer...). it's hardcoded for geoip fyi
