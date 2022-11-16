package stats

import (
	"encoding/json"
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

type summarizedTeamRoleStats struct {
	Attack  *genericStats
	Defence *genericStats
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
			var atkData, defData []ubiDetailedStatsJSON
			if err := json.Unmarshal(jsons[i].TeamRoles.Attack, &atkData); err != nil {
				return err
			}
			if err := json.Unmarshal(jsons[i].TeamRoles.Defence, &defData); err != nil {
				return err
			}
			*field = &summarizedTeamRoleStats{
				Attack:  newGenericStats(atkData[0]),
				Defence: newGenericStats(defData[0]),
			}
		}
	}
	return nil
}

type OperatorStats struct {
	abstractNamedStats
}

func (s *OperatorStats) AggregationType() string {
	return "operators"
}

type MapStats struct {
	abstractNamedStats
}

func (s *MapStats) AggregationType() string {
	return "maps"
}

type abstractNamedStats struct {
	statMetadata
	Casual   *namedTeamRoleStats
	Unranked *namedTeamRoleStats
	Ranked   *namedTeamRoleStats
}

type namedTeamRoleStats struct {
	Attack  []namedStats
	Defence []namedStats
}

type namedStats struct {
	Name string
	genericStats
}

func (s *abstractNamedStats) Load(metadata statMetadata, data *ubiGameModesJSON) error {
	s.statMetadata = metadata
	fields := []**namedTeamRoleStats{&s.Casual, &s.Unranked, &s.Ranked}
	jsons := []*ubiTeamRolesJSON{data.StatsCasual, data.StatsUnranked, data.StatsRanked}
	for i, field := range fields {
		if jsons[i] != nil {
			var atkData, defData []ubiDetailedStatsJSON
			if err := json.Unmarshal(jsons[i].TeamRoles.Attack, &atkData); err != nil {
				return err
			}
			if err := json.Unmarshal(jsons[i].TeamRoles.Defence, &defData); err != nil {
				return err
			}
			atk := make([]namedStats, len(atkData))
			def := make([]namedStats, len(defData))
			for j := range atk {
				atk[j] = *newOperatorTeamRoleStats(atkData[j])
			}
			for j := range def {
				def[j] = *newOperatorTeamRoleStats(defData[j])
			}

			*field = &namedTeamRoleStats{
				Attack:  atk,
				Defence: def,
			}
		}
	}
	return nil
}

func newOperatorTeamRoleStats(v ubiDetailedStatsJSON) *namedStats {
	var name string
	if v.StatsDetail == nil {
		name = "n/a"
	} else {
		name = *v.StatsDetail
	}
	return &namedStats{
		Name:         name,
		genericStats: *newGenericStats(v),
	}
}

type WeaponStats struct {
	statMetadata
	Casual   *weaponTeamRoleStats
	Unranked *weaponTeamRoleStats
	Ranked   *weaponTeamRoleStats
}

type weaponTeamRoleStats struct {
	Attack  weaponSlotStats
	Defence weaponSlotStats
}

type weaponSlotStats struct {
	Primary   weaponTypeMap
	Secondary weaponTypeMap
}

// weaponTypeMap maps weapon types to weapon stats
type weaponTypeMap map[string][]weaponStats

type weaponStats struct {
	WeaponName string
	reducedStats
}

func (s *WeaponStats) AggregationType() string {
	return "weapons"
}

func (s *WeaponStats) Load(metadata statMetadata, data *ubiGameModesJSON) error {
	s.statMetadata = metadata
	fields := []**weaponTeamRoleStats{&s.Casual, &s.Unranked, &s.Ranked}
	jsons := []*ubiTeamRolesJSON{data.StatsCasual, data.StatsUnranked, data.StatsRanked}
	for i, field := range fields {
		if jsons[i] != nil {
			var atkData, defData ubiWeaponSlotsJSON
			if err := json.Unmarshal(jsons[i].TeamRoles.Attack, &atkData); err != nil {
				return err
			}
			if err := json.Unmarshal(jsons[i].TeamRoles.Defence, &defData); err != nil {
				return err
			}
			atkPrimary := newWeaponTeamRoleStats(atkData.WeaponSlots.Primary)
			atkSecondary := newWeaponTeamRoleStats(atkData.WeaponSlots.Secondary)
			defPrimary := newWeaponTeamRoleStats(defData.WeaponSlots.Primary)
			defSecondary := newWeaponTeamRoleStats(defData.WeaponSlots.Secondary)

			*field = &weaponTeamRoleStats{
				Attack: weaponSlotStats{
					Primary:   atkPrimary,
					Secondary: atkSecondary,
				},
				Defence: weaponSlotStats{
					Primary:   defPrimary,
					Secondary: defSecondary,
				},
			}
		}
	}
	return nil
}

func newWeaponTeamRoleStats(v ubiWeaponTypesJSON) weaponTypeMap {
	result := weaponTypeMap{}
	for _, weaponType := range v.WeaponTypes {
		weapons := make([]weaponStats, len(weaponType.Weapons))
		for i, weapon := range weaponType.Weapons {
			weapons[i] = weaponStats{
				WeaponName: weapon.WeaponName,
				reducedStats: reducedStats{
					Headshots:           weapon.Headshots,
					HeadshotPercentage:  weapon.HeadshotPercentage,
					Kills:               weapon.Kills,
					RoundsPlayed:        weapon.RoundsPlayed,
					RoundsWon:           weapon.RoundsWon,
					RoundsLost:          weapon.RoundsLost,
					RoundsWithKill:      weapon.RoundsWithKill,
					RoundsWithMultikill: weapon.RoundsWithMultikill,
				},
			}
		}
		result[weaponType.WeaponTypeName] = weapons
	}
	return result
}

type reducedStats struct {
	Headshots           int
	HeadshotPercentage  float64
	Kills               int
	RoundsPlayed        int
	RoundsWon           int
	RoundsLost          int
	RoundsWithKill      float64
	RoundsWithMultikill float64
}

type genericStats struct {
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
	EntryDeaths          int
	EntryDeathTrades     int
	EntryKills           int
	EntryKillTrades      int
	Trades               int
	Revives              int
	RoundsSurvived       float64
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

func newGenericStats(v ubiDetailedStatsJSON) *genericStats {
	return &genericStats{
		reducedStats: reducedStats{
			Headshots:           v.Headshots,
			HeadshotPercentage:  v.HeadshotPercentage.Value,
			Kills:               v.Kills,
			RoundsPlayed:        v.RoundsPlayed,
			RoundsWon:           v.RoundsWon,
			RoundsLost:          v.RoundsLost,
			RoundsWithKill:      v.RoundsWithKill.Value,
			RoundsWithMultikill: v.RoundsWithMultikill.Value,
		},
		MatchesPlayed:        v.MatchesPlayed,
		MatchesWon:           v.MatchesWon,
		MatchesLost:          v.MatchesLost,
		MinutesPlayed:        v.MinutesPlayed,
		Assists:              v.Assists,
		Deaths:               v.Deaths,
		KillsPerRound:        v.KillsPerRound.Value,
		MeleeKills:           v.MeleeKills,
		TeamKills:            v.TeamKills,
		EntryDeaths:          v.EntryDeaths,
		EntryDeathTrades:     v.EntryDeathTrades,
		EntryKills:           v.EntryKills,
		EntryKillTrades:      v.EntryKillTrades,
		Trades:               v.Trades,
		Revives:              v.Revives,
		RoundsSurvived:       v.RoundsSurvived.Value,
		RoundsWithAce:        v.RoundsWithAce.Value,
		RoundsWithClutch:     v.RoundsWithClutch.Value,
		RoundsWithKOST:       v.RoundsWithKOST.Value,
		RoundsWithEntryDeath: v.RoundsWithEntryDeath.Value,
		RoundsWithEntryKill:  v.RoundsWithEntryKill.Value,
		DistancePerRound:     v.DistancePerRound,
		DistanceTotal:        v.DistanceTotal,
		TimeAlivePerMatch:    v.TimeAlivePerMatch,
		TimeDeadPerMatch:     v.TimeDeadPerMatch,
	}
}
