package user

type UserCollection struct {
	Wish    []Subject `json:"wish"`
	Collect []Subject `json:"collect"`
	Doing   []Subject `json:"doing"`
	OnHold  []Subject `json:"on_hold"`
	Dropped []Subject `json:"dropped"`
}

type Subject struct {
	SubjectID int      `json:"subject_id"`
	ProjectID int      `json:"project_id"`
	Tags      []string `json:"tags"`
	Comment   string   `json:"comment"`
	Rate      int      `json:"rate"`
	UpdatedAt string   `json:"updated_at"`
}

type JsonUserFile struct {
	UserID    int    `json:"user_id"`
	ProjectID int    `json:"project_id"`
	UserName  string `json:"name"`
	Data      UserCollection
}

type ApiResponse struct {
	Data   []Collection `json:"data"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

type Collection struct {
	UpdatedAt   string   `json:"updated_at"`
	Comment     string   `json:"comment"`
	Tags        []string `json:"tags"`
	SubjectID   int      `json:"subject_id"`
	SubjectType int      `json:"subject_type"`
	Type        int      `json:"type"`
	Rate        int      `json:"rate"`
}
