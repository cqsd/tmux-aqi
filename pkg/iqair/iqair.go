package iqair

import (
	"time"
)

type IQAirResponse struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Location struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Weather struct {
	Ts time.Time `json:"ts"`
	Tp int       `json:"tp"`
	Pr int       `json:"pr"`
	Hu int       `json:"hu"`
	Ws float64   `json:"ws"`
	Wd int       `json:"wd"`
	Ic string    `json:"ic"`
}

type Pollution struct {
	Ts     time.Time `json:"ts"`
	AqiUS  int       `json:"aqius"`
	MainUS string    `json:"mainus"`
	AqiCN  int       `json:"aqicn"`
	MainCN string    `json:"maincn"`
}

type Current struct {
	Weather   Weather   `json:"weather"`
	Pollution Pollution `json:"pollution"`
}

type Data struct {
	City     string   `json:"city"`
	State    string   `json:"state"`
	Country  string   `json:"country"`
	Location Location `json:"location"`
	Current  Current  `json:"current"`
}
