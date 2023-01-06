package stats

import (
	"encoding/json"
	"fmt"
	"text/template"
)

// TODO: test mapName=??
// TODO: test bombsite=??

var UbiStatsURLTemplate = template.Must(template.New("statsURL").Parse(
	"https://prod.datadev.ubisoft.com/v1/users/{{urlquery .ProfileID}}/playerstats?spaceId=5172a557-50b5-4665-b7db-e3f2e8c5041d&view={{urlquery .View}}&aggregation={{urlquery .Aggregation}}&gameMode=all,ranked,unranked,casual&platformGroup=PC&teamRole=all,Attacker,Defender&seasons={{urlquery .Season}}",
))

type UbiStatsURLParams struct {
	ProfileID   string
	Aggregation string
	View        string
	Season      string
}

type ubiStatsResponseJSON struct {
	ProfileData map[string]struct {
		Platforms struct {
			PC struct {
				GameModes ubiGameModesJSON `json:"gameModes"`
			} `json:"PC"`
		} `json:"platforms"`
	} `json:"profileData"`
	UserID string `json:"userId"`
}

type ubiGameModesJSON struct {
	StatsAll      *ubiTypedGameModeJSON `json:"all"`
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
	case typeTeamRoles, "":
		u.Value = new(ubiTeamRolesJSON)
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
type ubiTeamRolesJSON struct {
	TeamRoles struct {
		All     []ubiTypedTeamRoleJSON `json:"all"`
		Attack  []ubiTypedTeamRoleJSON `json:"Attacker"`
		Defence []ubiTypedTeamRoleJSON `json:"Defender"`
	} `json:"teamRoles"`
}

type teamRoleStatsType string

const (
	typeSeasonal    teamRoleStatsType = "Seasonal"
	typeMovingPoint teamRoleStatsType = "Moving Point Average Trend"
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
	case typeSeasonal:
		u.Value = new(ubiDetailedStatsJSON)
	case typeMovingPoint:
		u.Value = new(ubiMovingTrendJSON)
	default:
		return fmt.Errorf("encountered unknown team role type: '%s'", typed.Type)
	}
	err := json.Unmarshal(data, u.Value)
	u.ubiTeamRoleStatsTypeJSON = typed
	return err
}

/*
***************
Team Role Weapons
****************
*/
type ubiGameModeWeaponsJSON struct {
	TeamRoles struct {
		All     *ubiWeaponSlotsJSON `json:"all"`
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

/***************************
Moving Point Average (Trend)
***************************/

type ubiMovingTrendJSON struct {
	MovingPoints           int                     `json:"movingPoints"`
	DistancePerRound       ubiMovingTrendEntryJSON `json:"distancePerRound"`
	HeadshotPercentage     ubiMovingTrendEntryJSON `json:"headshotAccuracy"`
	KillDeathRatio         ubiMovingTrendEntryJSON `json:"killDeathRatio"`
	KillsPerRound          ubiMovingTrendEntryJSON `json:"killsPerRound"`
	RatioTimeAlivePerMatch ubiMovingTrendEntryJSON `json:"ratioTimeAlivePerMatch"`
	RoundsSurvived         ubiMovingTrendEntryJSON `json:"roundsSurvived"`
	RoundsWithKill         ubiMovingTrendEntryJSON `json:"roundsWithAKill"`
	RoundsWithKOST         ubiMovingTrendEntryJSON `json:"roundsWithKOST"`
	RoundsWithMultikill    ubiMovingTrendEntryJSON `json:"roundsWithMultiKill"`
	RoundsWithOpeningDeath ubiMovingTrendEntryJSON `json:"roundsWithOpeningDeath"`
	RoundsWithOpeningKill  ubiMovingTrendEntryJSON `json:"roundsWithOpeningKill"`
	WinLossRatio           ubiMovingTrendEntryJSON `json:"winLossRatio"`
}

type ubiMovingTrendEntryJSON struct {
	Low     float64              `json:"low"`
	Average float64              `json:"average"`
	High    float64              `json:"high"`
	Actuals ubiMovingTrendPoints `json:"actuals"`
	Trend   ubiMovingTrendPoints `json:"trend"`
}

type ubiMovingTrendPoints map[int]float64

/***************
Generic structs
****************/

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
	ubiSeasonInfo
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

type ubiSeasonInfo struct {
	SeasonYear   *string `json:"seasonYear"`
	SeasonNumber *string `json:"seasonNumber"`
}
