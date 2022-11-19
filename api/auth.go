package api

import (
	"encoding/json"
	"os"
	"time"
)

type ticket struct {
	Email      *string   `json:"email"`
	Name       string    `json:"nameOnPlatform"`
	ProfileID  string    `json:"profileId"`
	SessionID  string    `json:"sessionId"`
	Expiration time.Time `json:"expiration"`
	Token      string    `json:"ticket"`
}

func (t *ticket) IsExpired() bool {
	return time.Now().After(t.Expiration)
}

const ticketFile = "ticket.json"

func (t *ticket) Save() (err error) {
	var data []byte
	data, err = json.Marshal(t)
	if err != nil {
		return
	}

	var file *os.File
	file, err = os.OpenFile(ticketFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = file.Write(data)
	return
}

func canLoadTicket() (ok bool, err error) {
	_, statErr := os.Stat(ticketFile)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			ok = false
		} else {
			err = statErr
		}
	} else {
		ok = true
	}
	return
}

func loadTicket() (t *ticket, err error) {
	var file *os.File
	file, err = os.Open(ticketFile)
	if err != nil {
		return
	}

	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	t = new(ticket)
	err = json.NewDecoder(file).Decode(t)
	return
}
