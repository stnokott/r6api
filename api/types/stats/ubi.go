package stats

import (
	"encoding/json"
	"fmt"
	"text/template"
	"time"
)

// TODO: test view=seasonal

var UbiStatsURLTemplate = template.Must(template.New("statsURL").Parse(
	"https://prod.datadev.ubisoft.com/v1/profiles/{{urlquery .ProfileID}}/playerstats?spaceId=5172a557-50b5-4665-b7db-e3f2e8c5041d&view=current&aggregation={{urlquery .Aggregation}}&gameMode=ranked,unranked,casual&platform=PC&teamRole=attacker,defender&seasons={{urlquery .Season}}",
))

type UbiStatsURLParams struct {
	ProfileID   string
	Aggregation string
	Season      string
}

type ubiStatsResponseJSON struct {
	StartDate ubiTime `json:"startDate"`
	EndDate   ubiTime `json:"endDate"`
	Platforms struct {
		PC struct {
			GameModes ubiGameModesJSON `json:"gameModes"`
		} `json:"PC"`
	} `json:"platforms"`
}

type ubiGameModesJSON struct {
	StatsCasual   *ubiTypedGameModeJSON `json:"casual"`
	StatsUnranked *ubiTypedGameModeJSON `json:"unranked"`
	StatsRanked   *ubiTypedGameModeJSON `json:"ranked"`
}

/********************
Game Mode Stats Types
*********************/

type gameModeStatsType string

const (
	typeTeamRoles       gameModeStatsType = "Team roles"
	typeTeamRoleWeapons gameModeStatsType = "Team roles weapons"
)

type ubiGameModeStatsTypeJSON struct {
	Type gameModeStatsType `json:"type"`
}

type ubiTypedGameModeJSON struct {
	ubiGameModeStatsTypeJSON
	Value any
}

func (u *ubiTypedGameModeJSON) UnmarshalJSON(data []byte) error {
	var typed ubiGameModeStatsTypeJSON
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}

	switch typed.Type {
	case typeTeamRoles:
		u.Value = new(ubiGameModeJSON)
	case typeTeamRoleWeapons:
		u.Value = new(ubiGameModeWeaponsJSON)
	default:
		return fmt.Errorf("encountered unknown game mode type: '%s'", typed.Type)
	}
	err := json.Unmarshal(data, u.Value)
	u.Type = typed.Type
	return err
}

/********************
Team Roles Stats Types
*********************/

// ////////////
// Team Roles
// ////////////
type ubiGameModeJSON struct {
	TeamRoles struct {
		Attack  []ubiTypedTeamRoleJSON `json:"attacker"`
		Defence []ubiTypedTeamRoleJSON `json:"defender"`
	} `json:"teamRoles"`
}

// TODO: remove if only one type

type teamRoleStatsType string

const (
	typeGeneralized teamRoleStatsType = "Generalized"
)

type ubiTeamRoleStatsTypeJSON struct {
	Type      teamRoleStatsType `json:"type"`
	StatsType *string           `json:"statsType"`
}

type ubiTypedTeamRoleJSON struct {
	ubiTeamRoleStatsTypeJSON
	Value any
}

func (u *ubiTypedTeamRoleJSON) UnmarshalJSON(data []byte) error {
	var typed ubiTeamRoleStatsTypeJSON
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}

	switch typed.Type {
	case typeGeneralized:
		u.Value = new(ubiDetailedStatsJSON)
	default:
		return fmt.Errorf("encountered unknown team role type: '%s'", typed.Type)
	}
	err := json.Unmarshal(data, u.Value)
	u.ubiTeamRoleStatsTypeJSON = typed
	return err
}

// ///////////////////
// Team Role Weapons
// ///////////////////
type ubiGameModeWeaponsJSON struct {
	TeamRoles struct {
		Attack  *ubiWeaponSlotsJSON `json:"attacker"`
		Defence *ubiWeaponSlotsJSON `json:"defender"`
	} `json:"teamRoles"`
}

type ubiWeaponSlotsJSON struct {
	WeaponSlots struct {
		Primary   *ubiWeaponTypesJSON `json:"primaryWeapons"`
		Secondary *ubiWeaponTypesJSON `json:"secondaryWeapons"`
	} `json:"weaponSlots"`
}

type ubiWeaponTypesJSON struct {
	WeaponTypes []struct {
		WeaponTypeName string               `json:"weaponType"`
		Weapons        []ubiWeaponStatsJSON `json:"weapons"`
	} `json:"weaponTypes"`
}

type ubiWeaponStatsJSON struct {
	WeaponName string `json:"weaponName"`
	ubiReducedStatsJSON
	RoundsWithKill      float64 `json:"roundsWithAKill"`
	RoundsWithMultikill float64 `json:"roundsWithMultikill"`
	HeadshotPercentage  float64 `json:"headshotAccuracy"`
}

/***************
Generic structs
****************/

const ubiDateFormat = "20060102"

type ubiTime struct {
	time.Time
}

func (t *ubiTime) UnmarshalJSON(b []byte) (err error) {
	t.Time, err = time.Parse(ubiDateFormat, string(b))
	return
}

type ubiJSONFloat struct {
	Value float64 `json:"value"`
}

type ubiReducedStatsJSON struct {
	Headshots    int `json:"headshots"`
	Kills        int `json:"kills"`
	RoundsPlayed int `json:"roundsPlayed"`
	RoundsWon    int `json:"roundsWon"`
	RoundsLost   int `json:"roundsLost"`
}

type ubiDetailedStatsJSON struct {
	StatsDetail *string `json:"statsDetail"`
	ubiReducedStatsJSON
	MatchesPlayed        int          `json:"matchesPlayed"`
	MatchesWon           int          `json:"matchesWon"`
	MatchesLost          int          `json:"matchesLost"`
	MinutesPlayed        int          `json:"minutesPlayed"`
	Assists              int          `json:"assists"`
	Deaths               int          `json:"death"`
	KillsPerRound        ubiJSONFloat `json:"killsPerRound"`
	MeleeKills           int          `json:"meleeKills"`
	TeamKills            int          `json:"teamKills"`
	HeadshotPercentage   ubiJSONFloat `json:"headshotAccuracy"`
	EntryDeaths          int          `json:"openingDeaths"`
	EntryDeathTrades     int          `json:"openingDeathTrades"`
	EntryKills           int          `json:"openingKills"`
	EntryKillTrades      int          `json:"openingKillTrades"`
	Trades               int          `json:"trades"`
	Revives              int          `json:"revives"`
	RoundsSurvived       ubiJSONFloat `json:"roundsSurvived"`
	RoundsWithKill       ubiJSONFloat `json:"roundsWithAKill"`
	RoundsWithMultikill  ubiJSONFloat `json:"roundsWithMultikill"`
	RoundsWithAce        ubiJSONFloat `json:"roundsWithAce"`
	RoundsWithClutch     ubiJSONFloat `json:"roundsWithClutch"`
	RoundsWithKOST       ubiJSONFloat `json:"roundsWithKOST"`
	RoundsWithEntryDeath ubiJSONFloat `json:"roundsWithOpeningDeath"`
	RoundsWithEntryKill  ubiJSONFloat `json:"roundsWithOpeningKill"`
	DistancePerRound     float64      `json:"distancePerRound"`
	DistanceTotal        float64      `json:"distanceTravelled"`
	TimeAlivePerMatch    float64      `json:"timeAlivePerMatch"`
	TimeDeadPerMatch     float64      `json:"timeDeadPerMatch"`
}
