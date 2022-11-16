package skill

import (
	"text/template"
	"time"
)

// TODO: test view=seasonal

var UbiStatsURLTemplate = template.Must(template.New("statsURL").Parse(
	"https://prod.datadev.ubisoft.com/v1/profiles/{{.ProfileID}}/playerstats?spaceId=5172a557-50b5-4665-b7db-e3f2e8c5041d&view=current&aggregation={{.Aggregation}}&gameMode=ranked,unranked,casual&platform=PC&teamRole=attacker,defender&seasons={{.Season}}",
))

type UbiStatsURLParams struct {
	ProfileID   string
	Aggregation string
	Season      string
}

type UbiSkillRecordsJSON struct {
	SkillRecords []ubiSeasonSkillJSON `json:"seasons_player_skill_records"`
}

type ubiSeasonSkillJSON struct {
	SeasonID int `json:"season_id"`
	RegionSkills []ubiRegionSkillJSON `json:"regions_player_skill_records"`
}

type ubiRegionSkillJSON struct {
	RegionID string `json:"region_id`
	BoardSkills []ubiBoardSkillJSON `json:"boards_player_skill_records"`
}

type ubiBoardSkillJSON struct {
	BoardID string `json:"board_id"`
	PlayerSkills []ubiPlayerSkillJSON `json:"players_skill_records"`
}

type ubiPlayerSkillJSON struct {
	Abandons                  int       `json:"abandons"`
	Deaths                    int       `json:"deaths"`
	Kills                     int       `json:"kills"`
	LastMMRChange             int       `json:"last_match_mmr_change"`
	LastResult           int       `json:"last_match_result"`
	LastSkillMeanChange  int       `json:"last_match_skill_mean_change"`
	LastSkillStdevChange int       `json:"last_match_skill_stdev_change"`
	Losses                    int       `json:"losses"`
	MaxMMR                    int       `json:"max_mmr"`
	MaxRank                   int       `json:"max_rank"`
	MMR                       int       `json:"mmr"`
	NextRankMMR               int       `json:"next_rank_mmr"`
	PreviousRankMMR           int       `json:"previous_rank_mmr"`
	Rank                      int       `json:"rank"`
	Season                    int       `json:"season"`
	SkillMean                 int       `json:"skill_mean"`
	SkillStdev                float64   `json:"skill_stdev"`
	TopRankPosition           int       `json:"top_rank_position"`
	UpdateTime                time.Time `json:"update_time"`
	Wins                      int       `json:"wins"`
}