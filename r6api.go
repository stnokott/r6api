package r6api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/stnokott/r6api/auth"
	"github.com/stnokott/r6api/request"
	"github.com/stnokott/r6api/types/metadata"
	"github.com/stnokott/r6api/types/ranked"
	"github.com/stnokott/r6api/types/stats"

	"github.com/rs/zerolog"
)

type Profile struct {
	Name      string
	ProfileID string
}

func (p *Profile) ProfilePicURL() string {
	return fmt.Sprintf("https://ubisoft-avatars.akamaized.net/%s/default_146_146.png?appId=3587dcbb-7f81-457c-9781-0e3f29f6f56a", p.ProfileID)
}

func (p *Profile) MarshalZerologObject(e *zerolog.Event) {
	e.Str("username", p.Name).Str("profileID", p.ProfileID).Send()
	e.Discard()
}

type R6API struct {
	authCredentials string
	email           string
	ticket          *auth.Ticket
	logger          zerolog.Logger
}

// NewR6API creates a new instance with the provided login credentials and logger.
func NewR6API(email string, password string, logger zerolog.Logger) *R6API {
	authInput := []byte(email + ":" + password)
	authCredentials := base64.StdEncoding.EncodeToString(authInput)
	return &R6API{
		authCredentials: authCredentials,
		email:           email,
		ticket:          nil,
		logger:          logger,
	}
}

const ubiLoginRequestURL string = "https://public-ubiservices.ubi.com/v3/profiles/sessions"
const ubiAppIDAuth string = "39baebad-39e5-4552-8c25-2c9b919064e2"

// login performs login, caching the response ticket.
// It does this regardless of whether a ticket is already cached, so make sure to check before, e.g. with checkAuthentication().
func (a *R6API) login() (err error) {
	a.logger.Debug().Msg("attempting login")
	var body []byte
	body, err = json.Marshal(map[string]string{"rememberMe": "true"})
	if err != nil {
		return
	}

	req, _ := http.NewRequest("POST", ubiLoginRequestURL, bytes.NewBuffer(body))
	req.Header.Add("Ubi-AppId", ubiAppIDAuth)
	req.Header.Add("Authorization", "Basic "+a.authCredentials)
	req.Header.Add("Content-Type", "application/json")

	t := new(auth.Ticket)
	err = request.JSON(req, t)
	if err != nil {
		return
	}
	email := a.email
	t.Email = &email

	a.ticket = t
	a.logger.Info().Msgf("successfully logged in as <%s>", a.ticket.Name)
	err = a.ticket.Save()
	return
}

// checkAuthentication ensures the API contains an authorized, non-expired ticket by using the cached ticket or logging in again if non-existing or expired.
func (a *R6API) checkAuthentication() (err error) {
	loginReason := ""
	if a.ticket == nil {
		var canLoad bool
		canLoad, err = auth.CanLoadTicket()
		if err != nil {
			return
		}
		if canLoad {
			t, errL := auth.LoadTicket()
			if errL != nil {
				err = errL
				return
			}
			a.ticket = t
			a.logger.Debug().Msg("using cached ticket")
		} else {
			loginReason = "no cached token"
		}
	}
	if a.ticket != nil {
		if a.ticket.Email == nil || *a.ticket.Email != a.email {
			loginReason = "email mismatch"
		} else if a.ticket.IsExpired() {
			loginReason = "cached token expired"
		}
	}

	if loginReason != "" {
		a.logger.Debug().Msgf("login required, reason: %s", loginReason)
		err = a.login()
	}
	return
}

const ubiAppIDStats string = "3587dcbb-7f81-457c-9781-0e3f29f6f56a"

// requestAuthorized executes an authorized request (i.e. with the corresponding auth headers) and attempts to unmarshal the response into dst.
func (a *R6API) requestAuthorized(url string, dst any) (err error) {
	if err = a.checkAuthentication(); err != nil {
		return
	}
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Ubi-AppId", ubiAppIDStats)
	req.Header.Add("Ubi-SessionId", a.ticket.SessionID)
	req.Header.Add("Expiration", a.ticket.Expiration.Format("2006-01-02T15:04:05.999Z"))
	req.Header.Add("Authorization", "ubi_v1 t="+a.ticket.Token)

	err = request.JSON(req, dst)
	return
}

const ubiProfilesURLTemplate string = "https://public-ubiservices.ubi.com/v3/profiles?namesOnPlatform=%s&platformType=uplay"

type ubiProfileResp struct {
	Profiles []struct {
		Name      string `json:"nameOnPlatform"`
		ProfileID string `json:"profileId"`
	} `json:"profiles"`
}

// ResolveUser attempts to resolve the provided username to a Profile instance which can then be used for other requests.
func (a *R6API) ResolveUser(username string) (*Profile, error) {
	a.logger.Debug().Str("username", username).Msg("resolving profile")
	requestURL := fmt.Sprintf(ubiProfilesURLTemplate, url.QueryEscape(username))
	var p ubiProfileResp
	if err := a.requestAuthorized(requestURL, &p); err != nil {
		return nil, err
	}

	if len(p.Profiles) == 0 {
		return nil, fmt.Errorf("no user with name <%s> found", username)
	}
	resolvedName := p.Profiles[0].Name
	resolvedProfileID := p.Profiles[0].ProfileID
	if resolvedName != username {
		return nil, fmt.Errorf("no user with exact name <%s> found, closest match was <%s>", username, resolvedName)
	}
	a.logger.Debug().
		Str("username", username).
		Msgf("resolved to profile ID %s", resolvedProfileID)
	return &Profile{
		Name:      resolvedName,
		ProfileID: resolvedProfileID,
	}, nil
}

// GetMetadata retrieves information about seasons, i.e. season slug or MMR bounds.
// This is an expensive operation as it performs Javascript evaluations.
func (a *R6API) GetMetadata() (m *metadata.Metadata, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", metadata.URL, nil)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml")

	a.logger.Info().Msg("getting metadata")
	var body io.ReadCloser
	body, err = request.Plain(req)
	if err != nil {
		return
	}

	defer func() {
		errClose := body.Close()
		if err == nil {
			err = errClose
		}
	}()

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(body)
	if err != nil {
		return
	}
	m, err = metadata.New(doc.Find("script").Text())
	return
}

// GetStats retrieves statistics for a specific profile and season, loading the results into dst.
// dst needs to implement stats.Provider, the preconfigured providers can be found in the stats package.
func (a *R6API) GetStats(profile *Profile, season string, dst stats.Provider) error {
	args := stats.UbiStatsURLParams{
		ProfileID:   profile.ProfileID,
		Aggregation: dst.AggregationType(),
		View:        dst.ViewType(),
		Season:      season,
	}
	a.logger.Info().
		Str("username", profile.Name).
		Str("type", args.Aggregation).
		Str("season", season).
		Msg("getting stats")
	requestURLBytes := bytes.NewBuffer([]byte{})
	if err := stats.UbiStatsURLTemplate.Execute(requestURLBytes, args); err != nil {
		return err
	}

	if err := a.requestAuthorized(requestURLBytes.String(), dst); err != nil {
		return err
	}
	return nil
}

// GetRankedHistory returns a list of stats for the last numSeasons past ranked seasons.
// The resulting list will be ordered historically, i.e. the most-recent season last.
func (a *R6API) GetRankedHistory(profile *Profile, numSeasons uint8) (ranked.SkillHistory, error) {
	args := ranked.UbiSkillURLParams{
		ProfileID:      profile.ProfileID,
		NumPastSeasons: numSeasons,
	}
	a.logger.Info().
		Str("username", profile.Name).
		Uint8("seasons", numSeasons).
		Msg("getting ranked history")
	requestURLBytes := bytes.NewBuffer([]byte{})
	if err := ranked.UbiSkillURLTemplate.Execute(requestURLBytes, args); err != nil {
		return nil, err
	}

	resp := new(ranked.UbiSkillRecordsJSON)
	if err := a.requestAuthorized(requestURLBytes.String(), resp); err != nil {
		return nil, err
	}

	result, err := ranked.GetSkillHistory(resp)
	if err != nil {
		return nil, err
	}
	return result, nil
}
