package types

import (
	"fmt"
	"time"

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

func LoadStats[T StatsLoader](resp *UbiStatsResponseJSON, dst T) error {
	return dst.Load(
		statMetadata{
			resp.StartDate.Time,
			resp.EndDate.Time,
		},
		resp.Platforms.PC.GameModes,
	)
}

type StatsLoader interface {
	Load(metadata statMetadata, data *ubiGameModesJSON) error
	AggregationType() string
}

type statMetadata struct {
	RangeFrom time.Time
	RangeTo   time.Time
}

type SummarizedStats struct {
	statMetadata
	Casual   *summarizedTeamRoleStats
	Unranked *summarizedTeamRoleStats
	Ranked   *summarizedTeamRoleStats
}

func (s *SummarizedStats) AggregationType() string {
	return "summary"
}

func (s *SummarizedStats) Load(metadata statMetadata, data *ubiGameModesJSON) error {
	s.statMetadata = metadata
	fields := []**summarizedTeamRoleStats{&s.Casual, &s.Unranked, &s.Ranked}
	jsons := []*ubiTeamRolesJSON{data.StatsCasual, data.StatsUnranked, data.StatsRanked}
	for i, field := range fields {
		if jsons[i] != nil {
			*field = &summarizedTeamRoleStats{
				Attack:  newGenericStats(&jsons[i].TeamRoles.Attack[0]),
				Defence: newGenericStats(&jsons[i].TeamRoles.Defence[0]),
			}
		}
	}
	return nil
}

type OperatorStats struct {
	statMetadata
	Casual   *operatorTeamRoleStats
	Unranked *operatorTeamRoleStats
	Ranked   *operatorTeamRoleStats
}

func (s *OperatorStats) AggregationType() string {
	return "operators"
}

func (s *OperatorStats) Load(metadata statMetadata, data *ubiGameModesJSON) error {
	s.statMetadata = metadata
	fields := []**operatorTeamRoleStats{&s.Casual, &s.Unranked, &s.Ranked}
	jsons := []*ubiTeamRolesJSON{data.StatsCasual, data.StatsUnranked, data.StatsRanked}
	for i, field := range fields {
		if jsons[i] != nil {
			atk := make([]operatorStats, len(jsons[i].TeamRoles.Attack))
			def := make([]operatorStats, len(jsons[i].TeamRoles.Defence))
			for j := range atk {
				atk[j] = *newOperatorTeamRoleStats(&jsons[i].TeamRoles.Attack[j])
			}
			for j := range def {
				def[j] = *newOperatorTeamRoleStats(&jsons[i].TeamRoles.Defence[j])
			}

			*field = &operatorTeamRoleStats{
				Attack:  atk,
				Defence: def,
			}
		}
	}
	return nil
}

type summarizedTeamRoleStats struct {
	Attack  *genericStats
	Defence *genericStats
}

type operatorTeamRoleStats struct {
	Attack  []operatorStats
	Defence []operatorStats
}

type genericStats struct {
	MatchesPlayed          int
	MatchesWon             int
	MatchesLost            int
	RoundsPlayed           int
	RoundsWon              int
	RoundsLost             int
	MinutesPlayed          int
	Assists                int
	Deaths                 int
	Kills                  int
	KillsPerRound          float64
	Headshots              int
	HeadshotPercentage     float64
	MeleeKills             int
	TeamKills              int
	OpeningDeaths          int
	OpeningDeathTrades     int
	OpeningKills           int
	OpeningKillTrades      int
	Trades                 int
	Revives                int
	RoundsSurvived         float64
	RoundsWithKill         float64
	RoundsWithAce          float64
	RoundsWithClutch       float64
	RoundsWithKOST         float64
	RoundsWithMultikill    float64
	RoundsWithOpeningDeath float64
	RoundsWithOpeningKill  float64
	DistancePerRound       float64
	DistanceTotal          float64
	TimeAlivePerMatch      float64
	TimeDeadPerMatch       float64
}

func newGenericStats(v *ubiStatsJSON) *genericStats {
	return &genericStats{
		MatchesPlayed:          v.MatchesPlayed,
		MatchesWon:             v.MatchesWon,
		MatchesLost:            v.MatchesLost,
		RoundsPlayed:           v.RoundsPlayed,
		RoundsWon:              v.RoundsWon,
		RoundsLost:             v.RoundsLost,
		MinutesPlayed:          v.MinutesPlayed,
		Assists:                v.Assists,
		Deaths:                 v.Deaths,
		Kills:                  v.Kills,
		KillsPerRound:          v.KillsPerRound.Value,
		Headshots:              v.Headshots,
		HeadshotPercentage:     v.HeadshotPercentage.Value,
		MeleeKills:             v.MeleeKills,
		TeamKills:              v.TeamKills,
		OpeningDeaths:          v.OpeningDeaths,
		OpeningDeathTrades:     v.OpeningDeathTrades,
		OpeningKills:           v.OpeningKills,
		OpeningKillTrades:      v.OpeningKillTrades,
		Trades:                 v.Trades,
		Revives:                v.Revives,
		RoundsSurvived:         v.RoundsSurvived.Value,
		RoundsWithKill:         v.RoundsWithKill.Value,
		RoundsWithAce:          v.RoundsWithAce.Value,
		RoundsWithClutch:       v.RoundsWithClutch.Value,
		RoundsWithKOST:         v.RoundsWithKOST.Value,
		RoundsWithMultikill:    v.RoundsWithMultikill.Value,
		RoundsWithOpeningDeath: v.RoundsWithOpeningDeath.Value,
		RoundsWithOpeningKill:  v.RoundsWithOpeningKill.Value,
		DistancePerRound:       v.DistancePerRound,
		DistanceTotal:          v.DistanceTotal,
		TimeAlivePerMatch:      v.TimeAlivePerMatch,
		TimeDeadPerMatch:       v.TimeDeadPerMatch,
	}
}

type operatorStats struct {
	OperatorName string
	genericStats
}

func newOperatorTeamRoleStats(v *ubiStatsJSON) *operatorStats {
	var name string
	if v.Name == nil {
		name = "n/a"
	} else {
		name = *v.Name
	}
	return &operatorStats{
		OperatorName: name,
		genericStats: *newGenericStats(v),
	}
}
