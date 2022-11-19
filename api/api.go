package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/stnokott/r6api/api/types/ranked"
	"github.com/stnokott/r6api/api/types/stats"

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

type UbiAPI struct {
	authCredentials string
	email           string
	ticket          *ticket
	logger          zerolog.Logger
}

func NewUbiAPI(email string, password string, logger zerolog.Logger) *UbiAPI {
	authInput := []byte(email + ":" + password)
	authCredentials := base64.StdEncoding.EncodeToString(authInput)
	return &UbiAPI{
		authCredentials: authCredentials,
		email:           email,
		ticket:          nil,
		logger:          logger,
	}
}

const ubiLoginRequestURL string = "https://public-ubiservices.ubi.com/v3/profiles/sessions"
const ubiAppIDAuth string = "39baebad-39e5-4552-8c25-2c9b919064e2"

func (a *UbiAPI) Login() (err error) {
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

	t := new(ticket)
	err = request(req, t)
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

func (a *UbiAPI) checkAuthentication() (err error) {
	loginReason := ""
	if a.ticket == nil {
		var canLoad bool
		canLoad, err = canLoadTicket()
		if err != nil {
			return
		}
		if canLoad {
			t, errL := loadTicket()
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
		err = a.Login()
	}
	return
}

func (a *UbiAPI) requestAuthorized(url string, dst any) (err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Ubi-AppId", ubiAppIDStats)
	req.Header.Add("Ubi-SessionId", a.ticket.SessionID)
	req.Header.Add("Expiration", a.ticket.Expiration.Format("2006-01-02T15:04:05Z"))
	req.Header.Add("Authorization", "ubi_v1 t="+a.ticket.Token)

	err = request(req, dst)
	return
}

const ubiProfilesURLTemplate string = "https://public-ubiservices.ubi.com/v3/profiles?namesOnPlatform=%s&platformType=uplay"
const ubiAppIDStats string = "39baebad-39e5-4552-8c25-2c9b919064e2"

type ubiProfileResp struct {
	Profiles []struct {
		Name      string `json:"nameOnPlatform"`
		ProfileID string `json:"profileId"`
	} `json:"profiles"`
}

func (a *UbiAPI) ResolveUser(username string) (*Profile, error) {
	if err := a.checkAuthentication(); err != nil {
		return nil, err
	}
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

func (a *UbiAPI) GetStats(profile *Profile, season string, dst stats.Provider) error {
	args := stats.UbiStatsURLParams{
		ProfileID:   profile.ProfileID,
		Aggregation: dst.AggregationType(),
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

	a.logger.Info().Msg("...done")
	return nil
}

func (a *UbiAPI) GetRankedHistory(profile *Profile, numSeasons int8) (ranked.SkillHistory, error) {
	args := ranked.UbiSkillURLParams{
		ProfileID:      profile.ProfileID,
		NumPastSeasons: numSeasons,
	}
	a.logger.Info().
		Str("username", profile.Name).
		Int8("numSeasons", numSeasons).
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
	a.logger.Info().Msg("...done")
	return result, nil
}
