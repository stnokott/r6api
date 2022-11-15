package types

import "time"

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
		Attack  []ubiStatsJSON `json:"attacker"`
		Defence []ubiStatsJSON `json:"defender"`
	} `json:"teamRoles"`
}

type ubiJSONFloat struct {
	Value float64 `json:"value"`
}

type ubiStatsJSON struct {
	Name                   *string      `json:"statsDetail"`
	MatchesPlayed          int          `json:"matchesPlayed"`
	MatchesWon             int          `json:"matchesWon"`
	MatchesLost            int          `json:"matchesLost"`
	RoundsPlayed           int          `json:"roundsPlayed"`
	RoundsWon              int          `json:"roundsWon"`
	RoundsLost             int          `json:"roundsLost"`
	MinutesPlayed          int          `json:"minutesPlayed"`
	Assists                int          `json:"assists"`
	Deaths                 int          `json:"death"`
	Kills                  int          `json:"kills"`
	KillsPerRound          ubiJSONFloat `json:"killsPerRound"`
	Headshots              int          `json:"headshots"`
	HeadshotPercentage     ubiJSONFloat `json:"headshotAccuracy"`
	MeleeKills             int          `json:"meleeKills"`
	TeamKills              int          `json:"teamKills"`
	OpeningDeaths          int          `json:"openingDeaths"`
	OpeningDeathTrades     int          `json:"openingDeathTrades"`
	OpeningKills           int          `json:"openingKills"`
	OpeningKillTrades      int          `json:"openingKillTrades"`
	Trades                 int          `json:"trades"`
	Revives                int          `json:"revives"`
	RoundsSurvived         ubiJSONFloat `json:"roundsSurvived"`
	RoundsWithKill         ubiJSONFloat `json:"roundsWithKill"`
	RoundsWithAce          ubiJSONFloat `json:"roundsWithAce"`
	RoundsWithClutch       ubiJSONFloat `json:"roundsWithClutch"`
	RoundsWithKOST         ubiJSONFloat `json:"roundsWithKOST"`
	RoundsWithMultikill    ubiJSONFloat `json:"roundsWithMultikill"`
	RoundsWithOpeningDeath ubiJSONFloat `json:"roundsWithOpeningDeath"`
	RoundsWithOpeningKill  ubiJSONFloat `json:"roundsWithOpeningKill"`
	DistancePerRound       float64      `json:"distancePerRound"`
	DistanceTotal          float64      `json:"distanceTravelled"`
	TimeAlivePerMatch      float64      `json:"timeAlivePerMatch"`
	TimeDeadPerMatch       float64      `json:"timeDeadPerMatch"`
}
