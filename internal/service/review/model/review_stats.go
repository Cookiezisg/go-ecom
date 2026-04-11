package model

// ReviewStats 评价统计
type ReviewStats struct {
	TotalCount    int64   `json:"total_count"`
	Rating5Count  int64   `json:"rating_5_count"`
	Rating4Count  int64   `json:"rating_4_count"`
	Rating3Count  int64   `json:"rating_3_count"`
	Rating2Count  int64   `json:"rating_2_count"`
	Rating1Count  int64   `json:"rating_1_count"`
	AverageRating float64 `json:"average_rating"`
}
