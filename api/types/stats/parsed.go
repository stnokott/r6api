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
	TeamRoleType() gameModeStatsType
	LoadGameMode(GameMode, *ubiTypedGameModeJSON) error
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
		if jsn.Type != dst.TeamRoleType() {
			return fmt.Errorf("unexpected game mode stats type: '%s', expected '%s'", jsn.Type, dst.TeamRoleType())
		}
		if err = dst.LoadGameMode(fields[i], jsn); err != nil {
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

func (s *SummarizedStats) TeamRoleType() gameModeStatsType {
	return typeTeamRoles
}

func (s *SummarizedStats) UnmarshalJSON(data []byte) error {
	return unmarshalTeamRoleStats(s, data)
}

func (s *SummarizedStats) LoadGameMode(m GameMode, v *ubiTypedGameModeJSON) (err error) {
	jsn, ok := v.Value.(*ubiGameModeJSON)
	if !ok {
		err = fmt.Errorf("could not cast json (%s) to required struct (*%s)", reflect.TypeOf(v.Value), reflect.TypeOf(ubiGameModeJSON{}))
		return
	}
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
	teamRole := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	fields := []**detailedStats{&stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			continue
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
		*fields[i] = newDetailedTeamRoleStats(data)
	}
	return
}

func newDetailedTeamRoleStats(data *ubiDetailedStatsJSON) *detailedStats {
	return &detailedStats{
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

func (s *OperatorStats) TeamRoleType() gameModeStatsType {
	return typeTeamRoles
}

func (s *OperatorStats) UnmarshalJSON(data []byte) (err error) {
	return unmarshalTeamRoleStats(s, data)
}

func (s *OperatorStats) LoadGameMode(m GameMode, v *ubiTypedGameModeJSON) (err error) {
	jsn, ok := v.Value.(*ubiGameModeJSON)
	if !ok {
		err = fmt.Errorf("could not cast json (%s) to required struct (*%s)", reflect.TypeOf(v.Value), reflect.TypeOf(ubiGameModeJSON{}))
		return
	}
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

	teamRole := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	resultFields := []*[]operatorTeamRoleStats{&stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			continue
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
				OperatorName:  operatorName,
				detailedStats: *newDetailedTeamRoleStats(data),
			}
		}
		*resultFields[i] = resultTeamRoleData
	}
	return
}

/*
Weapons structs
*/

type WeaponStats struct {
	statsMetadata
	Casual   *weaponTeamRoles
	Unranked *weaponTeamRoles
	Ranked   *weaponTeamRoles
}

type weaponTeamRoles struct {
	Attack  *weaponTypes
	Defence *weaponTypes
}

type weaponTypes struct {
	PrimaryWeapons   weaponTypesMap
	SecondaryWeapons weaponTypesMap
}

type weaponTypesMap map[string][]weaponNamedStats

type weaponNamedStats struct {
	WeaponName string
	reducedStats
	RoundsWithKill      float64
	RoundsWithMultikill float64
	HeadshotPercentage  float64
}

func (s *WeaponStats) AggregationType() string {
	return "weapons"
}

func (s *WeaponStats) TeamRoleType() gameModeStatsType {
	return typeTeamRoleWeapons
}

func (s *WeaponStats) UnmarshalJSON(data []byte) (err error) {
	return unmarshalTeamRoleStats(s, data)
}

func (s *WeaponStats) LoadGameMode(m GameMode, v *ubiTypedGameModeJSON) (err error) {
	jsn, ok := v.Value.(*ubiGameModeWeaponsJSON)
	if !ok {
		err = fmt.Errorf("could not cast json (%s) to required struct (*%s)", reflect.TypeOf(v.Value), reflect.TypeOf(ubiGameModeJSON{}))
		return
	}
	stats := new(weaponTeamRoles)
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

	inputTeamRoles := []*ubiWeaponSlotsJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**weaponTypes{&stats.Attack, &stats.Defence}

	for i, inputTeamRole := range inputTeamRoles {
		if inputTeamRole == nil {
			continue
		}
		outputTeamRoleData := new(weaponTypes)
		inputWeaponSlots := []*ubiWeaponTypesJSON{inputTeamRole.WeaponSlots.Primary, inputTeamRole.WeaponSlots.Secondary}
		outputWeaponSlots := []*weaponTypesMap{&outputTeamRoleData.PrimaryWeapons, &outputTeamRoleData.SecondaryWeapons}
		for j, inputWeaponSlot := range inputWeaponSlots {
			if inputWeaponSlot == nil {
				continue
			}
			*outputWeaponSlots[j] = newWeaponTypesMap(inputWeaponSlot)
		}
		*outputTeamRoles[i] = outputTeamRoleData
	}
	return
}

func newWeaponTypesMap(v *ubiWeaponTypesJSON) weaponTypesMap {
	result := weaponTypesMap{}
	for _, weaponType := range v.WeaponTypes {
		weaponTypeStats := make([]weaponNamedStats, len(weaponType.Weapons))
		for j, weaponStats := range weaponType.Weapons {
			weaponTypeStats[j] = weaponNamedStats{
				WeaponName: weaponStats.WeaponName,
				reducedStats: reducedStats{
					Headshots:    weaponStats.Headshots,
					Kills:        weaponStats.Kills,
					RoundsPlayed: weaponStats.RoundsPlayed,
					RoundsWon:    weaponStats.RoundsWon,
					RoundsLost:   weaponStats.RoundsLost,
				},
				RoundsWithKill:      weaponStats.RoundsWithKill,
				RoundsWithMultikill: weaponStats.RoundsWithMultikill,
				HeadshotPercentage:  weaponStats.HeadshotPercentage,
			}
		}
		result[weaponType.WeaponTypeName] = weaponTypeStats
	}
	return result
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
