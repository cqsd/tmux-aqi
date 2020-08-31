package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/cqsd/tmux-aqi/pkg/iqair"
)

var configDirName string = ".iq-air"

// fileExists returns true if a file exists and is not a dir
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// dirExists returns true if a file exists and is a dir
func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// joinConfigDir returns an absolute path inside the config dir
func joinConfigDir(path string) string {
	usr, _ := user.Current()
	home := usr.HomeDir
	return filepath.Join(home, configDirName, path)
}

// getKey looks for an IQAir API key and returns it. env IQAIR_API_KEY or
// ~/.iq-air/key. env overrides config file
func getKey() (string, error) {
	key := os.Getenv("IQAIR_API_KEY")
	if key != "" {
		return key, nil
	} else {
		key, err := ioutil.ReadFile(joinConfigDir("key"))
		if err != nil {
			return "", fmt.Errorf("no api key found. set IQAIR_API_KEY or put it in ~/.iq-air/key")
		}
		return strings.TrimSpace(string(key)), nil
	}
}

// previousRun is a wrapper around the iqair response with a timestamp of when
// the result was retrieved.  need this because iqair timestamp is the time of
// last _update_, which isn't useful for rate limit stuff
type previousRun struct {
	Ts   time.Time            `json:"ts"`
	Data *iqair.IQAirResponse `json:"data"`
}

// since iqair doesn't actually update very quickly, check cache first to
// decide if we need to make a new call. returns nil if previous run is stale
func getPreviousRun(maxAge time.Duration) *iqair.IQAirResponse {
	lastrunFile := joinConfigDir("lastrun.json")
	if !fileExists(lastrunFile) {
		return nil
	}

	// try using it to prevent further errors.
	lastrun, err := ioutil.ReadFile(lastrunFile)
	previous := previousRun{}
	err = json.Unmarshal(lastrun, &previous)
	if err != nil {
		return nil
	}

	// check if it's been at least 5 minutes since last run
	loc, _ := time.LoadLocation("UTC")
	expiry := previous.Ts.In(loc).Add(maxAge)
	if time.Now().In(loc).After(expiry) {
		// current time UTC is past expiration
		return nil
	}

	// last run is still fresh
	return previous.Data
}

// cache the latest run
func writeNewRun(data *iqair.IQAirResponse) error {
	lastrunFile := joinConfigDir("lastrun.json")

	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)
	currentRun := previousRun{Ts: now, Data: data}
	b, err := json.Marshal(&currentRun)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(lastrunFile, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

// format the iqair response as a tmux string with colors
func toTmuxString(data *iqair.IQAirResponse) string {
	city := data.Data.City
	aqi := data.Data.Current.Pollution.AqiUS
	var fg, bg string
	if aqi <= 50 {
		fg = "yellow"
		bg = "green"
	} else if aqi <= 100 {
		fg = "brightwhite"
		bg = "red"
	} else {
		fg = "brightwhite"
		bg = "purple"
	}
	return fmt.Sprintf("#[fg=%s,bg=%s] %s AQI: %d ", fg, bg, city, aqi)
}

func checkOrExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// create the config dir if it doesn't exist
	configDirPath := joinConfigDir("")
	if !dirExists(configDirPath) {
		err := os.Mkdir(configDirPath, 0755)
		checkOrExit(err)
	}

	data := getPreviousRun(time.Minute * 5)
	if data != nil {
		// have a cached result that's still fresh
		fmt.Println(toTmuxString(data))
	} else {
		query := url.Values{}
		key, err := getKey()
		checkOrExit(err)

		query.Add("key", key)
		u, _ := url.Parse("http://api.airvisual.com/v2/nearest_city")
		u.RawQuery = query.Encode()
		res, err := http.Get(u.String())
		checkOrExit(err)

		body, err := ioutil.ReadAll(res.Body)
		checkOrExit(err)

		data := iqair.IQAirResponse{}
		err = json.Unmarshal(body, &data)
		checkOrExit(err)

		fmt.Println(toTmuxString(&data))

		// cache the latest run
		writeNewRun(&data)
		checkOrExit(err)
	}
}
