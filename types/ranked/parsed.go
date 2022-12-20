package ranked

import (
	"errors"
	"fmt"
	"time"
)

// GetSkillHistory attempts to parse v into a SkillHistory instance.
func GetSkillHistory(v *UbiSkillRecordsJSON) (SkillHistory, error) {
	history := make(SkillHistory, len(v.SkillRecords))
	for i, record := range v.SkillRecords {
		if len(record.RegionSkills) == 0 {
			return nil, fmt.Errorf("no regions found in response, expected '%s'", ubiRegionIDParam)
		}
		region := record.RegionSkills[0]
		if region.RegionID != ubiRegionIDParam {
			return nil, fmt.Errorf("expected region '%s', got '%s'", ubiRegionIDParam, region.RegionID)
		}
		if len(region.BoardSkills) == 0 {
			return nil, fmt.Errorf("no boards found in response, expected '%s'", ubiBoardIDParam)
		}
		board := region.BoardSkills[0]
		if board.BoardID != ubiBoardIDParam {
			return nil, fmt.Errorf("expected board '%s', got '%s'", ubiBoardIDParam, board.BoardID)
		}
		if len(board.PlayerSkills) == 0 {
			return nil, errors.New("no skill reports found")
		}
		skill := board.PlayerSkills[0]
		history[i] = NewSeasonStats(skill)
	}
	return history, nil
}

// SkillHistory contains a list of season stats.
// Should be ordered historically (i.e. most-recent season last).
type SkillHistory []*SeasonStats

type SeasonStats struct {
	SeasonID             int
	Abandons             int
	Deaths               int
	Kills                int
	LastMMRChange        int
	LastResult           int
	LastSkillMeanChange  float64
	LastSkillStdevChange float64
	Losses               int
	MaxRank              int
	MaxMMR               int
	MMR                  int
	NextRankMMR          int
	PreviousRankMMR      int
	Rank                 int
	SkillMean            float64
	SkillStdev           float64
	TopRankPosition      int
	UpdateTime           time.Time
	Wins                 int
}

func NewSeasonStats(v ubiPlayerSkillJSON) *SeasonStats {
	return &SeasonStats{
		SeasonID:             v.Season,
		Abandons:             v.Abandons,
		Deaths:               v.Deaths,
		Kills:                v.Kills,
		LastMMRChange:        int(v.LastMMRChange),
		LastResult:           v.LastResult,
		LastSkillMeanChange:  v.LastSkillMeanChange,
		LastSkillStdevChange: v.LastSkillStdevChange,
		Losses:               v.Losses,
		MaxMMR:               int(v.MaxMMR),
		MaxRank:              v.MaxRank,
		MMR:                  int(v.MMR),
		NextRankMMR:          int(v.NextRankMMR),
		PreviousRankMMR:      int(v.PreviousRankMMR),
		Rank:                 v.Rank,
		SkillMean:            v.SkillMean,
		SkillStdev:           v.SkillStdev,
		TopRankPosition:      v.TopRankPosition,
		UpdateTime:           v.UpdateTime,
		Wins:                 v.Wins,
	}
}
