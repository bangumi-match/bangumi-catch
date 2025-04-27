package subject

type FileCollection struct {
	Collect int `json:"collect"`
	Doing   int `json:"doing"`
	Dropped int `json:"dropped"`
	OnHold  int `json:"on_hold"`
	Wish    int `json:"wish"`
}

type Images struct {
	Common string `json:"common"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Small  string `json:"small"`
}

type Infobox struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type RatingCount struct {
	One   int `json:"1"`
	Ten   int `json:"10"`
	Two   int `json:"2"`
	Three int `json:"3"`
	Four  int `json:"4"`
	Five  int `json:"5"`
	Six   int `json:"6"`
	Eight int `json:"8"`
	Nine  int `json:"9"`
}

type Rating struct {
	Count RatingCount `json:"count"`
	Rank  int         `json:"rank"`
	Score float64     `json:"score"`
	Total int         `json:"total"`
}

type FileTag struct {
	Count     int    `json:"count"`
	Name      string `json:"name"`
	TotalCont int    `json:"total_cont"`
}

type JsonSubject struct {
	Collection    FileCollection `json:"collection"`
	Date          string         `json:"date"`
	Eps           int            `json:"eps"`
	OriginalID    int            `json:"id"`
	Images        Images         `json:"images"`
	Infobox       []Infobox      `json:"infobox"`
	Locked        bool           `json:"locked"`
	MetaTags      []interface{}  `json:"meta_tags"`
	Name          string         `json:"name"`
	NameCn        string         `json:"name_cn"`
	Nsfw          bool           `json:"nsfw"`
	Platform      string         `json:"platform"`
	Rating        Rating         `json:"rating"`
	Series        bool           `json:"series"`
	Summary       string         `json:"summary"`
	Tags          []FileTag      `json:"tags"`
	TotalEpisodes int            `json:"total_episodes"`
	Type          int            `json:"type"`
	Volumes       int            `json:"volumes"`
	ProjectID     int            `json:"project_id"`
}

type JsonSubjectPerson struct {
	Images   Images   `json:"images"`
	Name     string   `json:"name"`
	Relation string   `json:"relation"`
	Career   []string `json:"career"`
	Type     int      `json:"type"`
	ID       int      `json:"id"`
	Eps      string   `json:"eps"`
}

type JsonSubjectPersonCollection struct {
	JsonSubjectPersons []JsonSubjectPerson `json:"persons"`
	ProjectID          int                 `json:"project_id"`
	OriginalID         int                 `json:"id"`
}

type JsonSubjectRelation struct {
	Images   Images `json:"images"`
	Name     string `json:"name"`
	NameCn   string `json:"name_cn"`
	Relation string `json:"relation"`
	Type     int    `json:"type"`
	ID       int    `json:"id"`
}

type JsonSubjectRelationCollection struct {
	JsonSubjectRelations []JsonSubjectRelation `json:"relations"`
	ProjectID            int                   `json:"project_id"`
	OriginalID           int                   `json:"id"`
}
