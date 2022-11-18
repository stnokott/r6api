package stats

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type GameMode string

const (
	CASUAL   GameMode = "casual"
	UNRANKED GameMode = "unranked"
	RANKED   GameMode = "ranked"
)

type Provider interface {
	json.Unmarshaler
	AggregationType() string
	LoadGameMode(GameMode, *ubiGameModeJSON) error
}
type statsMetadata struct {
	TimeFrom time.Time
	TimeTo   time.Time
}

func unmarshalTeamRoleStats[T Provider](dst T, data []byte) (err error) {
	var raw ubiStatsResponseJSON
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}
	root := raw.Platforms.PC.GameModes
	if root.StatsCasual == nil && root.StatsUnranked == nil && root.StatsRanked == nil {
		return
	}
	jsons := []*ubiTypedGameModeJSON{root.StatsCasual, root.StatsUnranked, root.StatsRanked}
	fields := []GameMode{CASUAL, UNRANKED, RANKED}
	for i, jsn := range jsons {
		if jsn == nil {
			continue
		}
		if jsn.Type != typeTeamRoles {
			return fmt.Errorf("unexpected game mode stats type: '%s', expected '%s'", jsn.Type, typeTeamRoles)
		}
		teamRoles, ok := jsn.Value.(*ubiGameModeJSON)
		if !ok {
			return fmt.Errorf(
				"game mode data (%s) could not be cast to required struct (*%s)",
				reflect.TypeOf(jsn.Value).Name(),
				reflect.TypeOf(ubiGameModeJSON{}).Name(),
			)
		}
		if err = dst.LoadGameMode(fields[i], teamRoles); err != nil {
			return
		}
	}
	return
}

/***************
Summarized stats
 ***************/

type SummarizedStats struct {
	statsMetadata
	Casual   *summarizedGameModeStats
	Unranked *summarizedGameModeStats
	Ranked   *summarizedGameModeStats
}

type summarizedGameModeStats struct {
	Attack  *detailedStats
	Defence *detailedStats
}

func (s *SummarizedStats) AggregationType() string {
	return "summary"
}

func (s *SummarizedStats) UnmarshalJSON(data []byte) error {
	return unmarshalTeamRoleStats(s, data)
}

func (s *SummarizedStats) LoadGameMode(m GameMode, v *ubiGameModeJSON) (err error) {
	stats := new(summarizedGameModeStats)
	switch m {
	case CASUAL:
		s.Casual = stats
	case UNRANKED:
		s.Unranked = stats
	case RANKED:
		s.Ranked = stats
	default:
		err = fmt.Errorf("got invalid game mode: %s", m)
		return
	}
	teamRole := [][]ubiTypedTeamRoleJSON{v.TeamRoles.Attack, v.TeamRoles.Defence}
	fields := []**detailedStats{&stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			err = fmt.Errorf("no data found in team role field %s", reflect.TypeOf(teamRoleData).Name())
			return
		}
		data, ok := teamRoleData[0].Value.(*ubiDetailedStatsJSON)
		if !ok {
			err = fmt.Errorf(
				"team role data (%s) could not be cast to required struct (%s)",
				reflect.TypeOf(teamRoleData[0].Value).Name(),
				reflect.TypeOf(ubiDetailedStatsJSON{}).Name(),
			)
			return
		}
		*fields[i] = &detailedStats{
			reducedStats: reducedStats{
				Headshots:    data.Headshots,
				Kills:        data.Kills,
				RoundsPlayed: data.RoundsPlayed,
				RoundsWon:    data.RoundsWon,
				RoundsLost:   data.RoundsLost,
			},
			MatchesPlayed:        data.MatchesPlayed,
			MatchesWon:           data.MatchesWon,
			MatchesLost:          data.MatchesLost,
			MinutesPlayed:        data.MinutesPlayed,
			Assists:              data.Assists,
			Deaths:               data.Deaths,
			KillsPerRound:        data.KillsPerRound.Value,
			MeleeKills:           data.MeleeKills,
			TeamKills:            data.TeamKills,
			HeadshotPercentage:   data.HeadshotPercentage.Value,
			EntryDeaths:          data.EntryDeaths,
			EntryDeathTrades:     data.EntryDeathTrades,
			EntryKills:           data.EntryKills,
			EntryKillTrades:      data.EntryKillTrades,
			Trades:               data.Trades,
			Revives:              data.Revives,
			RoundsSurvived:       data.RoundsSurvived.Value,
			RoundsWithKill:       data.RoundsWithKill.Value,
			RoundsWithMultikill:  data.RoundsWithMultikill.Value,
			RoundsWithAce:        data.RoundsWithAce.Value,
			RoundsWithClutch:     data.RoundsWithClutch.Value,
			RoundsWithKOST:       data.RoundsWithKOST.Value,
			RoundsWithEntryDeath: data.RoundsWithEntryDeath.Value,
			RoundsWithEntryKill:  data.RoundsWithEntryKill.Value,
			DistancePerRound:     data.DistancePerRound,
			DistanceTotal:        data.DistanceTotal,
			TimeAlivePerMatch:    data.TimeAlivePerMatch,
			TimeDeadPerMatch:     data.TimeDeadPerMatch,
		}
	}
	return
}

/*************
Operator stats
 *************/

type OperatorStats struct {
	statsMetadata
	Casual   *operatorTeamRoles
	Unranked *operatorTeamRoles
	Ranked   *operatorTeamRoles
}

type operatorTeamRoles struct {
	Attack  []operatorTeamRoleStats
	Defence []operatorTeamRoleStats
}

type operatorTeamRoleStats struct {
	OperatorName string
	detailedStats
}

func (s *OperatorStats) AggregationType() string {
	return "operators"
}

func (s *OperatorStats) UnmarshalJSON(data []byte) (err error) {
	return unmarshalTeamRoleStats(s, data)
}

func (s *OperatorStats) LoadGameMode(m GameMode, v *ubiGameModeJSON) (err error) {
	stats := new(operatorTeamRoles)
	switch m {
	case CASUAL:
		s.Casual = stats
	case UNRANKED:
		s.Unranked = stats
	case RANKED:
		s.Ranked = stats
	default:
		err = fmt.Errorf("got invalid game mode: %s", m)
		return
	}

	teamRole := [][]ubiTypedTeamRoleJSON{v.TeamRoles.Attack, v.TeamRoles.Defence}
	resultFields := []*[]operatorTeamRoleStats{&stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			err = fmt.Errorf("no data found in team role field %s", reflect.TypeOf(teamRoleData).Name())
			return
		}
		resultTeamRoleData := make([]operatorTeamRoleStats, len(teamRoleData))
		for j, teamRoleStats := range teamRoleData {
			data, ok := teamRoleStats.Value.(*ubiDetailedStatsJSON)
			if !ok {
				err = fmt.Errorf(
					"team role data (%s) could not be cast to required struct (%s)",
					reflect.TypeOf(teamRoleData[0].Value).Name(),
					reflect.TypeOf(ubiDetailedStatsJSON{}).Name(),
				)
				return
			}
			var operatorName string
			if data.StatsDetail == nil {
				operatorName = "n/a"
			} else {
				operatorName = *data.StatsDetail
			}
			resultTeamRoleData[j] = operatorTeamRoleStats{
				OperatorName: operatorName,
				detailedStats: detailedStats{
					reducedStats: reducedStats{
						Headshots:    data.Headshots,
						Kills:        data.Kills,
						RoundsPlayed: data.RoundsPlayed,
						RoundsWon:    data.RoundsWon,
						RoundsLost:   data.RoundsLost,
					},
					MatchesPlayed:        data.MatchesPlayed,
					MatchesWon:           data.MatchesWon,
					MatchesLost:          data.MatchesLost,
					MinutesPlayed:        data.MinutesPlayed,
					Assists:              data.Assists,
					Deaths:               data.Deaths,
					KillsPerRound:        data.KillsPerRound.Value,
					MeleeKills:           data.MeleeKills,
					TeamKills:            data.TeamKills,
					HeadshotPercentage:   data.HeadshotPercentage.Value,
					EntryDeaths:          data.EntryDeaths,
					EntryDeathTrades:     data.EntryDeathTrades,
					EntryKills:           data.EntryKills,
					EntryKillTrades:      data.EntryKillTrades,
					Trades:               data.Trades,
					Revives:              data.Revives,
					RoundsSurvived:       data.RoundsSurvived.Value,
					RoundsWithKill:       data.RoundsWithKill.Value,
					RoundsWithMultikill:  data.RoundsWithMultikill.Value,
					RoundsWithAce:        data.RoundsWithAce.Value,
					RoundsWithClutch:     data.RoundsWithClutch.Value,
					RoundsWithKOST:       data.RoundsWithKOST.Value,
					RoundsWithEntryDeath: data.RoundsWithEntryDeath.Value,
					RoundsWithEntryKill:  data.RoundsWithEntryKill.Value,
					DistancePerRound:     data.DistancePerRound,
					DistanceTotal:        data.DistanceTotal,
					TimeAlivePerMatch:    data.TimeAlivePerMatch,
					TimeDeadPerMatch:     data.TimeDeadPerMatch,
				},
			}
		}
		*resultFields[i] = resultTeamRoleData
	}
	return
}

/*
Generic structs
*/

type reducedStats struct {
	Headshots    int
	Kills        int
	RoundsPlayed int
	RoundsWon    int
	RoundsLost   int
}

type detailedStats struct {
	reducedStats
	MatchesPlayed        int
	MatchesWon           int
	MatchesLost          int
	MinutesPlayed        int
	Assists              int
	Deaths               int
	KillsPerRound        float64
	MeleeKills           int
	TeamKills            int
	HeadshotPercentage   float64
	EntryDeaths          int
	EntryDeathTrades     int
	EntryKills           int
	EntryKillTrades      int
	Trades               int
	Revives              int
	RoundsSurvived       float64
	RoundsWithKill       float64
	RoundsWithMultikill  float64
	RoundsWithAce        float64
	RoundsWithClutch     float64
	RoundsWithKOST       float64
	RoundsWithEntryDeath float64
	RoundsWithEntryKill  float64
	DistancePerRound     float64
	DistanceTotal        float64
	TimeAlivePerMatch    float64
	TimeDeadPerMatch     float64
}
