package models

type Stats struct {
	Users        int `json:"users"`
	Applications int `json:"applications"`
	Approved     int `json:"approved"`
}
