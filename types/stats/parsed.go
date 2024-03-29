package stats

import (
	"encoding/json"
	"errors"
	"fmt"
)

type GameMode string

const (
	ALL      GameMode = "all"
	CASUAL   GameMode = "casual"
	UNRANKED GameMode = "unranked"
	RANKED   GameMode = "ranked"
)

// Provider should be implemented by statistics structs to enable it to be unmarshalled properly into the corresponding struct.
type Provider interface {
	json.Unmarshaler
	AggregationType() string // type of aggregation (e.g. "operators") to be used in URL query
	ViewType() string        // type of view (e.g. "summary") to be used in URL query
}

type statsLoader[TGameMode any, TJSON any] struct {
	All      *TGameMode
	Casual   *TGameMode
	Unranked *TGameMode
	Ranked   *TGameMode
}

func (l *statsLoader[TGameMode, TJSON]) loadRawStats(data []byte, dst Provider, loadTeamRoles func(*TJSON, *TGameMode) error) (err error) {
	var raw ubiStatsResponseJSON
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}
	root := raw.ProfileData[raw.UserID].Platforms.PC.GameModes

	gameModeJSONs := []*ubiTypedGameModeJSON{root.StatsAll, root.StatsCasual, root.StatsUnranked, root.StatsRanked}
	gameModes := []GameMode{ALL, CASUAL, UNRANKED, RANKED}
	for i, gameModeJSON := range gameModeJSONs {
		if gameModeJSON == nil {
			continue
		}
		jsn, ok := gameModeJSON.Value.(*TJSON)
		if !ok {
			return fmt.Errorf("could not cast json (%T) to required struct (*%T)", gameModeJSON.Value, *new(TJSON))
		}
		stats := new(TGameMode)
		switch gameModes[i] {
		case ALL:
			l.All = stats
		case CASUAL:
			l.Casual = stats
		case UNRANKED:
			l.Unranked = stats
		case RANKED:
			l.Ranked = stats
		default:
			return fmt.Errorf("got invalid game mode: %s", gameModes[i])
		}
		if err = loadTeamRoles(jsn, stats); err != nil {
			err = fmt.Errorf("could not load team roles for game mode %s: %w", gameModes[i], err)
			return
		}
	}
	return
}

/***************
Summarized stats
 ***************/

// SummarizedStats provides stats without any specific aggregation.
type SummarizedStats struct {
	statsLoader[SummarizedGameModeStats, ubiTeamRolesJSON]
}

type SummarizedGameModeStats struct {
	All     *DetailedStats
	Attack  *DetailedStats
	Defence *DetailedStats
	matchStats
}

func (s *SummarizedStats) AggregationType() string {
	return "summary"
}

func (s *SummarizedStats) ViewType() string {
	return "seasonal"
}

func (s *SummarizedStats) UnmarshalJSON(data []byte) error {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

func (s *SummarizedStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *SummarizedGameModeStats) (err error) {
	inputTeamRoles := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.All, jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**DetailedStats{&stats.All, &stats.Attack, &stats.Defence}

	for i, inputTeamRole := range inputTeamRoles {
		if len(inputTeamRole) == 0 {
			continue
		}
		data, ok := inputTeamRole[0].Value.(*ubiDetailedStatsJSON)
		if !ok {
			err = fmt.Errorf(
				"team role data (%T) could not be cast to required struct (%T)",
				inputTeamRole[0].Value,
				ubiDetailedStatsJSON{},
			)
			return
		}
		*outputTeamRoles[i] = newDetailedStats(data)

		if stats.matchStats.MatchesPlayed == 0 {
			stats.matchStats = newMatchStats(data)
		}
	}
	return
}

/*************
Operator stats
**************/

// OperatorStats provides stats aggregated by operator.
type OperatorStats struct {
	NamedStats
}

func (s *OperatorStats) AggregationType() string {
	return "operators"
}

func (s *OperatorStats) ViewType() string {
	return "seasonal"
}

func (s *OperatorStats) UnmarshalJSON(data []byte) error {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

/********
Map stats
*********/

// MapStats provides stats aggregated by map.
type MapStats struct {
	statsLoader[map[string]NamedMapStatDetails, ubiTeamRolesJSON]
}

type NamedMapStatDetails struct {
	matchStats
	DetailedStats
	Bombsites *BombsiteGamemodeStats
}

func (s *MapStats) AggregationType() string {
	return "maps"
}

func (s *MapStats) ViewType() string {
	return "current"
}

func (s *MapStats) UnmarshalJSON(data []byte) (err error) {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

func (s *MapStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *map[string]NamedMapStatDetails) (err error) {
	inputTeamRole := jsn.TeamRoles.All

	if len(inputTeamRole) == 0 {
		return errors.New("no input data for team role 'ALL'")
	}
	mapStats := map[string]NamedMapStatDetails{}
	for _, mapData := range inputTeamRole {
		data, ok := mapData.Value.(*ubiDetailedStatsJSON)
		if !ok {
			err = fmt.Errorf(
				"team role data (%T) could not be cast to required struct (%T)",
				mapData.Value,
				ubiDetailedStatsJSON{},
			)
			return
		}
		mapStats[*data.StatsDetail] = NamedMapStatDetails{
			DetailedStats: *newDetailedStats(data),
			matchStats:    newMatchStats(data),
		}
	}
	*stats = mapStats

	return
}

/*************
Bombsite stats
**************/

// BombsiteStats provides stats aggregated by map.
type BombsiteStats struct {
	statsLoader[BombsiteGamemodeStats, ubiTeamRolesJSON]
}

type BombsiteGamemodeStats struct {
	All     []BombsiteTeamRoleStats
	Attack  []BombsiteTeamRoleStats
	Defence []BombsiteTeamRoleStats
}

type BombsiteTeamRoleStats struct {
	DetailedStats
	Name string
}

func (s *BombsiteStats) AggregationType() string {
	return "bombsites"
}

func (s *BombsiteStats) ViewType() string {
	return "current"
}

func (s *BombsiteStats) UnmarshalJSON(data []byte) (err error) {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

func (s *BombsiteStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *BombsiteGamemodeStats) (err error) {
	inputTeamRoles := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.All, jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []*[]BombsiteTeamRoleStats{&stats.All, &stats.Attack, &stats.Defence}

	for i, inputTeamRole := range inputTeamRoles {
		if inputTeamRole == nil {
			continue
		}
		outputTeamRoleData := make([]BombsiteTeamRoleStats, len(inputTeamRole))

		for j, inputData := range inputTeamRole {
			data, ok := inputData.Value.(*ubiDetailedStatsJSON)
			if !ok {
				err = fmt.Errorf(
					"team role data (%T) could not be cast to required struct (%T)",
					inputData.Value,
					ubiDetailedStatsJSON{},
				)
				return
			}
			outputTeamRoleData[j] = BombsiteTeamRoleStats{
				DetailedStats: *newDetailedStats(data),
				Name:          *data.StatsDetail,
			}
		}

		*outputTeamRoles[i] = outputTeamRoleData
	}
	return
}

/**************
Weapons structs
***************/

// WeaponStats provides stats aggregated by weapon type and name.
type WeaponStats struct {
	statsLoader[WeaponTeamRoles, ubiGameModeWeaponsJSON]
}

type WeaponTeamRoles struct {
	All     *WeaponTypes
	Attack  *WeaponTypes
	Defence *WeaponTypes
}

type WeaponTypes struct {
	PrimaryWeapons   WeaponTypesMap
	SecondaryWeapons WeaponTypesMap
}

type WeaponTypesMap map[string]WeaponNamesMap

type WeaponNamesMap map[string]WeaponNamedStats

type WeaponNamedStats struct {
	reducedStats
	RoundsWithKill      float64
	RoundsWithMultikill float64
	HeadshotPercentage  float64
}

func (s *WeaponStats) AggregationType() string {
	return "weapons"
}

func (s *WeaponStats) ViewType() string {
	return "current" // does not seem to support seasonal view
}

func (s *WeaponStats) UnmarshalJSON(data []byte) error {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

func (s *WeaponStats) loadTeamRole(jsn *ubiGameModeWeaponsJSON, stats *WeaponTeamRoles) (err error) {
	inputTeamRoles := []*ubiWeaponSlotsJSON{jsn.TeamRoles.All, jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**WeaponTypes{&stats.All, &stats.Attack, &stats.Defence}

	for i, inputTeamRole := range inputTeamRoles {
		if inputTeamRole == nil {
			continue
		}
		outputTeamRoleData := new(WeaponTypes)
		inputWeaponSlots := []*ubiWeaponTypesJSON{inputTeamRole.WeaponSlots.Primary, inputTeamRole.WeaponSlots.Secondary}
		outputWeaponSlots := []*WeaponTypesMap{&outputTeamRoleData.PrimaryWeapons, &outputTeamRoleData.SecondaryWeapons}
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

func newWeaponTypesMap(v *ubiWeaponTypesJSON) WeaponTypesMap {
	result := WeaponTypesMap{}
	for _, weaponType := range v.WeaponTypes {
		weaponTypeStats := make(WeaponNamesMap, len(weaponType.Weapons))
		for _, weaponStats := range weaponType.Weapons {
			weaponTypeStats[weaponStats.WeaponName] = WeaponNamedStats{
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

/***************************
Moving Point Average (Trend)
***************************/

// MovingTrendStats provides stats without any specific aggregation, but with trends across a specific timeframe.
type MovingTrendStats struct {
	statsLoader[MovingTrendTeamRoles, ubiTeamRolesJSON]
}

type MovingTrendTeamRoles struct {
	All     *MovingTrend
	Attack  *MovingTrend
	Defence *MovingTrend
}

type MovingTrend struct {
	MovingPoints           int
	DistancePerRound       MovingTrendEntry
	HeadshotPercentage     MovingTrendEntry
	KillDeathRatio         MovingTrendEntry
	KillsPerRound          MovingTrendEntry
	RatioTimeAlivePerMatch MovingTrendEntry
	RoundsSurvived         MovingTrendEntry
	RoundsWithKill         MovingTrendEntry
	RoundsWithKOST         MovingTrendEntry
	RoundsWithMultikill    MovingTrendEntry
	RoundsWithOpeningDeath MovingTrendEntry
	RoundsWithOpeningKill  MovingTrendEntry
	WinLossRatio           MovingTrendEntry
}

type MovingTrendEntry struct {
	Low     float64
	Average float64
	High    float64
	Actuals MovingTrendPoints
	Trend   MovingTrendPoints
}

type MovingTrendPoints []float64

func (s *MovingTrendStats) AggregationType() string {
	return "movingpoint"
}

func (s *MovingTrendStats) ViewType() string {
	return "current" // does not seem to support seasonal
}

func (s *MovingTrendStats) UnmarshalJSON(data []byte) error {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

func (*MovingTrendStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *MovingTrendTeamRoles) (err error) {
	inputTeamRoles := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.All, jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**MovingTrend{&stats.All, &stats.Attack, &stats.Defence}

	for i, teamRole := range inputTeamRoles {
		if len(teamRole) == 0 {
			continue
		}
		data, ok := teamRole[0].Value.(*ubiMovingTrendJSON)
		if !ok {
			err = fmt.Errorf(
				"team role data (%T) could not be cast to required struct (%T)",
				teamRole[0].Value,
				ubiMovingTrendJSON{},
			)
			return
		}
		*outputTeamRoles[i] = newMovingTrendStats(data)
	}
	return
}

func newMovingTrendStats(v *ubiMovingTrendJSON) *MovingTrend {
	return &MovingTrend{
		MovingPoints:           v.MovingPoints,
		DistancePerRound:       newMovingTrendEntry(v.DistancePerRound),
		HeadshotPercentage:     newMovingTrendEntry(v.HeadshotPercentage),
		KillDeathRatio:         newMovingTrendEntry(v.KillDeathRatio),
		KillsPerRound:          newMovingTrendEntry(v.KillsPerRound),
		RatioTimeAlivePerMatch: newMovingTrendEntry(v.RatioTimeAlivePerMatch),
		RoundsSurvived:         newMovingTrendEntry(v.RoundsSurvived),
		RoundsWithKill:         newMovingTrendEntry(v.RoundsWithKill),
		RoundsWithKOST:         newMovingTrendEntry(v.RoundsWithKOST),
		RoundsWithMultikill:    newMovingTrendEntry(v.RoundsWithMultikill),
		RoundsWithOpeningDeath: newMovingTrendEntry(v.RoundsWithOpeningDeath),
		RoundsWithOpeningKill:  newMovingTrendEntry(v.RoundsWithOpeningKill),
		WinLossRatio:           newMovingTrendEntry(v.WinLossRatio),
	}
}

func newMovingTrendEntry(v ubiMovingTrendEntryJSON) MovingTrendEntry {
	result := MovingTrendEntry{
		Low:     v.Low,
		Average: v.Average,
		High:    v.High,
	}
	inputFields := []ubiMovingTrendPoints{v.Actuals, v.Trend}
	outputFields := []*MovingTrendPoints{&result.Actuals, &result.Trend}

	for i, inputField := range inputFields {
		points := make(MovingTrendPoints, len(inputField))
		for j, v := range inputField {
			points[j-1] = v // -1 since index in JSON starts at 1
		}
		*outputFields[i] = points
	}
	return result
}

/**************
Generic structs
***************/

type reducedStats struct {
	Headshots    int
	Kills        int
	RoundsPlayed int
	RoundsWon    int
	RoundsLost   int
}

type matchStats struct {
	MatchesPlayed int
	MatchesWon    int
	MatchesLost   int
}

type DetailedStats struct {
	reducedStats
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

func newDetailedStats(data *ubiDetailedStatsJSON) *DetailedStats {
	return &DetailedStats{
		reducedStats: reducedStats{
			Headshots:    data.Headshots,
			Kills:        data.Kills,
			RoundsPlayed: data.RoundsPlayed,
			RoundsWon:    data.RoundsWon,
			RoundsLost:   data.RoundsLost,
		},
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

func newTotalDetailedTeamRoleStats(data []ubiTypedTeamRoleJSON) (*DetailedStats, error) {
	v := &DetailedStats{
		reducedStats: reducedStats{
			Headshots:    0,
			Kills:        0,
			RoundsPlayed: 0,
			RoundsWon:    0,
			RoundsLost:   0,
		},
		MinutesPlayed:        0,
		Assists:              0,
		Deaths:               0,
		KillsPerRound:        0,
		MeleeKills:           0,
		TeamKills:            0,
		HeadshotPercentage:   0,
		EntryDeaths:          0,
		EntryDeathTrades:     0,
		EntryKills:           0,
		EntryKillTrades:      0,
		Trades:               0,
		Revives:              0,
		RoundsSurvived:       0,
		RoundsWithKill:       0,
		RoundsWithMultikill:  0,
		RoundsWithAce:        0,
		RoundsWithClutch:     0,
		RoundsWithKOST:       0,
		RoundsWithEntryDeath: 0,
		RoundsWithEntryKill:  0,
		DistancePerRound:     0,
		DistanceTotal:        0,
		TimeAlivePerMatch:    0,
		TimeDeadPerMatch:     0,
	}

	count := float64(len(data))
	for _, entry := range data {
		casted, ok := entry.Value.(*ubiDetailedStatsJSON)
		if !ok {
			return nil, fmt.Errorf(
				"team role data (%T) could not be cast to required struct (%T)",
				entry.Value,
				ubiDetailedStatsJSON{},
			)
		}
		v.Headshots += casted.Headshots
		v.Kills += casted.Kills
		v.RoundsPlayed += casted.RoundsPlayed
		v.RoundsWon += casted.RoundsWon
		v.RoundsLost += casted.RoundsLost
		v.MinutesPlayed += casted.MinutesPlayed
		v.Assists += casted.Assists
		v.Deaths += casted.Deaths
		v.KillsPerRound += casted.KillsPerRound.Value
		v.MeleeKills += casted.MeleeKills
		v.TeamKills += casted.TeamKills
		v.HeadshotPercentage += casted.HeadshotPercentage.Value
		v.EntryDeaths += casted.EntryDeaths
		v.EntryDeathTrades += casted.EntryDeathTrades
		v.EntryKills += casted.EntryKills
		v.EntryKillTrades += casted.EntryKillTrades
		v.Trades += casted.Trades
		v.Revives += casted.Revives
		v.RoundsSurvived += casted.RoundsSurvived.Value
		v.RoundsWithKill += casted.RoundsWithKill.Value
		v.RoundsWithMultikill += casted.RoundsWithKill.Value
		v.RoundsWithAce += casted.RoundsWithAce.Value
		v.RoundsWithClutch += casted.RoundsWithClutch.Value
		v.RoundsWithKOST += casted.RoundsWithKOST.Value
		v.RoundsWithEntryDeath += casted.RoundsWithEntryDeath.Value
		v.RoundsWithEntryKill += casted.RoundsWithEntryKill.Value
		v.DistancePerRound += casted.DistancePerRound
		v.DistanceTotal += casted.DistanceTotal
		v.TimeAlivePerMatch += casted.TimeAlivePerMatch
		v.TimeDeadPerMatch += casted.TimeDeadPerMatch
	}

	v.KillsPerRound /= count
	v.HeadshotPercentage /= count
	v.RoundsSurvived /= count
	v.RoundsWithKill /= count
	v.RoundsWithMultikill /= count
	v.RoundsWithAce /= count
	v.RoundsWithClutch /= count
	v.RoundsWithKOST /= count
	v.RoundsWithEntryDeath /= count
	v.RoundsWithEntryKill /= count
	v.DistancePerRound /= count
	v.TimeAlivePerMatch /= count
	v.TimeDeadPerMatch /= count
	return v, nil
}

type NamedStats struct {
	statsLoader[NamedTeamRoles, ubiTeamRolesJSON]
}

type NamedTeamRoleStats map[string]DetailedStats

type NamedTeamRoles struct {
	All     NamedTeamRoleStats
	Attack  NamedTeamRoleStats
	Defence NamedTeamRoleStats
}

func (s *NamedStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *NamedTeamRoles) (err error) {
	teamRole := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.All, jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	resultFields := []*NamedTeamRoleStats{&stats.All, &stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			continue
		}
		resultTeamRoleData := make(NamedTeamRoleStats, len(teamRoleData)+1)

		// calculate total
		var totalData *DetailedStats
		totalData, err = newTotalDetailedTeamRoleStats(teamRoleData)
		if err != nil {
			return
		}
		resultTeamRoleData["All"] = *totalData

		// calculate named
		for _, teamRoleStats := range teamRoleData {
			data, ok := teamRoleStats.Value.(*ubiDetailedStatsJSON)
			if !ok {
				err = fmt.Errorf(
					"team role data (%T) could not be cast to required struct (%T)",
					teamRoleData[0].Value,
					ubiDetailedStatsJSON{},
				)
				return
			}
			var name string
			if data.StatsDetail == nil {
				name = "n/a"
			} else {
				name = *data.StatsDetail
			}
			resultTeamRoleData[name] = *newDetailedStats(data)
		}
		*resultFields[i] = resultTeamRoleData
	}

	return
}

func newMatchStats(jsn *ubiDetailedStatsJSON) matchStats {
	return matchStats{
		MatchesPlayed: jsn.MatchesPlayed,
		MatchesWon:    jsn.MatchesWon,
		MatchesLost:   jsn.MatchesLost,
	}
}
