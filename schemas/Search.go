package schemas

import "time"

type SearchResultRespData struct {
	PostID         string    `json:"postId"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	Rank           float64   `json:"rank"`
}

type SearchRespData struct {
	Results []SearchResultRespData `json:"results"`
}
