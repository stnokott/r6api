package auth

import (
	"encoding/json"
	"os"
	"time"
)

type Ticket struct {
	Email      *string   `json:"email"`
	Name       string    `json:"nameOnPlatform"`
	ProfileID  string    `json:"profileId"`
	SessionID  string    `json:"sessionId"`
	Expiration time.Time `json:"expiration"`
	Token      string    `json:"ticket"`
}

// IsExpired checks if this ticket expired by comparing its expiration with the current time.
func (t *Ticket) IsExpired() bool {
	return time.Now().Add(5 * time.Minute).After(t.Expiration)
}

const ticketFile = "ticket.json"

// Save serializes this ticket to a file which can be loaded by Load().
// This is a means of caching, attempting to minimize authentication overhead for every request to the API.
func (t *Ticket) Save() (err error) {
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

// CanLoadTicket returns a boolean indicating if a cached ticket is present and can be loaded.
// Returns an error if an unexpected error occurs.
func CanLoadTicket() (ok bool, err error) {
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

// LoadTicket deserializes the cached ticket file and returns it.
// Should be prefaced with calling CanLoadTicket().
func LoadTicket() (t *Ticket, err error) {
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

	t = new(Ticket)
	err = json.NewDecoder(file).Decode(t)
	return
}
