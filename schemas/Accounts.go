package schemas

import "time"

type UserRespData struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PostRespData struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"authorId"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListPostsRespData struct {
	Posts []PostRespData `json:"posts"`
}
