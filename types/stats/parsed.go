package stats

import (
	"encoding/json"
	"fmt"
	"time"
)

type GameMode string

const (
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

type StatsMetadata struct {
	TimeFrom time.Time
	TimeTo   time.Time
}

type statsLoader[TGameMode any, TJSON any] struct {
	Casual   *TGameMode
	Unranked *TGameMode
	Ranked   *TGameMode
}

func (l *statsLoader[TGameMode, TJSON]) loadRawStats(data []byte, dst Provider, loadTeamRoles func(*TJSON, *TGameMode) error) (err error) {
	var raw ubiStatsResponseJSON
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}
	root := raw.Platforms.PC.GameModes
	if root.StatsCasual == nil && root.StatsUnranked == nil && root.StatsRanked == nil {
		return
	}
	gameModeJSONs := []*ubiTypedGameModeJSON{root.StatsCasual, root.StatsUnranked, root.StatsRanked}
	gameModes := []GameMode{CASUAL, UNRANKED, RANKED}
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
		case CASUAL:
			l.Casual = stats
		case UNRANKED:
			l.Unranked = stats
		case RANKED:
			l.Ranked = stats
		default:
			return fmt.Errorf("got invalid game mode: %s", gameModes[i])
		}
		return loadTeamRoles(jsn, stats)
	}
	return
}

/***************
Summarized stats
 ***************/

// SummarizedStats provides stats without any specific aggregation.
type SummarizedStats struct {
	statsLoader[SummarizedGameModeStats, ubiTeamRolesJSON]
	SeasonSlug string
	StatsMetadata
}

type SummarizedGameModeStats struct {
	Attack  *detailedStats
	Defence *detailedStats
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
	inputTeamRoles := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**detailedStats{&stats.Attack, &stats.Defence}

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
		*outputTeamRoles[i] = newDetailedTeamRoleStats(data)
		if s.SeasonSlug == "" {
			s.SeasonSlug = assembleSeasonSlug(data.ubiSeasonInfo)
		}
	}
	return
}

/*************
Operator stats
**************/

// OperatorStats provides stats aggregated by operator.
type OperatorStats struct {
	abstractNamedStats
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
	abstractNamedStats
}

func (s *MapStats) AggregationType() string {
	return "maps"
}

func (s *MapStats) ViewType() string {
	return "seasonal"
}

func (s *MapStats) UnmarshalJSON(data []byte) error {
	return s.loadRawStats(data, s, s.loadTeamRole)
}

/**************
Weapons structs
***************/

// WeaponStats provides stats aggregated by weapon type and name.
type WeaponStats struct {
	statsLoader[WeaponTeamRoles, ubiGameModeWeaponsJSON]
	StatsMetadata
}

type WeaponTeamRoles struct {
	Attack  *WeaponTypes
	Defence *WeaponTypes
}

type WeaponTypes struct {
	PrimaryWeapons   WeaponTypesMap
	SecondaryWeapons WeaponTypesMap
}

type WeaponTypesMap map[string][]WeaponNamedStats

type WeaponNamedStats struct {
	WeaponName string
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

func (*WeaponStats) loadTeamRole(jsn *ubiGameModeWeaponsJSON, stats *WeaponTeamRoles) (err error) {
	inputTeamRoles := []*ubiWeaponSlotsJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**WeaponTypes{&stats.Attack, &stats.Defence}

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
		weaponTypeStats := make([]WeaponNamedStats, len(weaponType.Weapons))
		for j, weaponStats := range weaponType.Weapons {
			weaponTypeStats[j] = WeaponNamedStats{
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

/***************************
Moving Point Average (Trend)
***************************/

// MovingTrendStats provides stats without any specific aggregation, but with trends across a specific timeframe.
type MovingTrendStats struct {
	statsLoader[MovingTrendTeamRoles, ubiTeamRolesJSON]
	StatsMetadata
	Casual   *MovingTrendTeamRoles
	Unranked *MovingTrendTeamRoles
	Ranked   *MovingTrendTeamRoles
}

type MovingTrendTeamRoles struct {
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
	inputTeamRoles := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	outputTeamRoles := []**MovingTrend{&stats.Attack, &stats.Defence}

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

func assembleSeasonSlug(v ubiSeasonInfo) string {
	year := "??"
	number := "??"
	if v.SeasonYear != nil {
		year = *v.SeasonYear
	}
	if v.SeasonNumber != nil {
		number = *v.SeasonNumber
	}
	return year + number
}

type abstractNamedStats struct {
	statsLoader[abstractNamedTeamRoles, ubiTeamRolesJSON]
	SeasonSlug string
	StatsMetadata
}

type abstractNamedTeamRoles struct {
	Attack  []abstractNamedTeamRoleStats
	Defence []abstractNamedTeamRoleStats
}

type abstractNamedTeamRoleStats struct {
	Name string
	detailedStats
}

func (s *abstractNamedStats) loadTeamRole(jsn *ubiTeamRolesJSON, stats *abstractNamedTeamRoles) (err error) {
	teamRole := [][]ubiTypedTeamRoleJSON{jsn.TeamRoles.Attack, jsn.TeamRoles.Defence}
	resultFields := []*[]abstractNamedTeamRoleStats{&stats.Attack, &stats.Defence}

	for i, teamRoleData := range teamRole {
		if len(teamRoleData) == 0 {
			continue
		}
		resultTeamRoleData := make([]abstractNamedTeamRoleStats, len(teamRoleData))
		for j, teamRoleStats := range teamRoleData {
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
			resultTeamRoleData[j] = abstractNamedTeamRoleStats{
				Name:          name,
				detailedStats: *newDetailedTeamRoleStats(data),
			}
			if s.SeasonSlug == "" {
				s.SeasonSlug = assembleSeasonSlug(data.ubiSeasonInfo)
			}
		}
		*resultFields[i] = resultTeamRoleData
	}
	return
}
