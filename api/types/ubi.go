package types

import (
	"encoding/json"
	"time"
)

const ubiDateFormat = "20060102"

type ubiTime struct {
	time.Time
}

func (t *ubiTime) UnmarshalJSON(b []byte) (err error) {
	t.Time, err = time.Parse(ubiDateFormat, string(b))
	return
}

type UbiStatsResponseJSON struct {
	StartDate ubiTime `json:"startDate"`
	EndDate   ubiTime `json:"endDate"`
	Platforms struct {
		PC struct {
			GameModes *ubiGameModesJSON `json:"gameModes"`
		} `json:"PC"`
	} `json:"platforms"`
}

type ubiGameModesJSON struct {
	StatsCasual   *ubiTeamRolesJSON `json:"casual"`
	StatsUnranked *ubiTeamRolesJSON `json:"unranked"`
	StatsRanked   *ubiTeamRolesJSON `json:"ranked"`
}

type ubiTeamRolesJSON struct {
	TeamRoles struct {
		Attack  json.RawMessage `json:"attacker"`
		Defence json.RawMessage `json:"defender"`
	} `json:"teamRoles"`
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

type ubiWeaponSlotsJSON struct {
	WeaponSlots struct {
		Primary   ubiWeaponTypesJSON `json:"primaryWeapons"`
		Secondary ubiWeaponTypesJSON `json:"secondaryWeapons"`
	} `json:"weaponSlots"`
}

type ubiWeaponTypesJSON struct {
	WeaponTypes []struct {
		WeaponTypeName string `json:"weaponType"`
		Weapons        []struct {
			WeaponName string `json:"weaponName"`
			ubiWeaponStatsJSON
		} `json:"weapons"`
	} `json:"weaponTypes"`
}

type ubiWeaponStatsJSON struct {
	ubiReducedStatsJSON
	RoundsWithKill      float64 `json:"roundsWithAKill"`
	RoundsWithMultikill float64 `json:"roundsWithMultikill"`
	HeadshotPercentage  float64 `json:"headshotAccuracy"`
}
