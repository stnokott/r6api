package ranked

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	ubiBoardIDParam  string = "pvp_ranked"
	ubiRegionIDParam string = "ncsa"
)

var UbiSkillURLTemplate = template.Must(template.New("skillURL").Parse(
	fmt.Sprintf(
		"https://public-ubiservices.ubi.com/v1/spaces/5172a557-50b5-4665-b7db-e3f2e8c5041d/sandboxes/OSBOR_PC_LNCH_A/r6karma/player_skill_records?board_ids=%s&season_ids={{.SeasonIDs}}&region_ids=%s&profile_ids={{.ProfileID}}",
		ubiBoardIDParam,
		ubiRegionIDParam,
	),
))

type UbiSkillURLParams struct {
	ProfileID      string
	NumPastSeasons int8
}

func (p UbiSkillURLParams) SeasonIDs() string {
	seasonIDs := make([]string, p.NumPastSeasons)
	for i := range seasonIDs {
		seasonIDs[i] = strconv.Itoa(-(i + 1))
	}
	return strings.Join(seasonIDs, ",")
}

type UbiSkillRecordsJSON struct {
	SkillRecords []ubiSeasonSkillJSON `json:"seasons_player_skill_records"`
}

type ubiSeasonSkillJSON struct {
	SeasonID     int                  `json:"season_id"`
	RegionSkills []ubiRegionSkillJSON `json:"regions_player_skill_records"`
}

type ubiRegionSkillJSON struct {
	RegionID    string              `json:"region_id"`
	BoardSkills []ubiBoardSkillJSON `json:"boards_player_skill_records"`
}

type ubiBoardSkillJSON struct {
	BoardID      string               `json:"board_id"`
	PlayerSkills []ubiPlayerSkillJSON `json:"players_skill_records"`
}

type ubiPlayerSkillJSON struct {
	Abandons             int       `json:"abandons"`
	Deaths               int       `json:"deaths"`
	Kills                int       `json:"kills"`
	LastMMRChange        float64   `json:"last_match_mmr_change"`
	LastResult           int       `json:"last_match_result"`
	LastSkillMeanChange  float64   `json:"last_match_skill_mean_change"`
	LastSkillStdevChange float64   `json:"last_match_skill_stdev_change"`
	Losses               int       `json:"losses"`
	MaxRank              int       `json:"max_rank"`
	MaxMMR               float64   `json:"max_mmr"`
	MMR                  float64   `json:"mmr"`
	NextRankMMR          float64   `json:"next_rank_mmr"`
	PreviousRankMMR      float64   `json:"previous_rank_mmr"`
	Rank                 int       `json:"rank"`
	Season               int       `json:"season"`
	SkillMean            float64   `json:"skill_mean"`
	SkillStdev           float64   `json:"skill_stdev"`
	TopRankPosition      int       `json:"top_rank_position"`
	UpdateTime           time.Time `json:"update_time"`
	Wins                 int       `json:"wins"`
}
