package conversion

// VideoInfo ...
type VideoInfo struct {
	From          string    `json:"From"`
	Uncensored    bool      `json:"Uncensored"`
	ID            string    `json:"ID"`
	Title         string    `json:"Title"`
	OriginalTitle string    `json:"OriginalTitle"`
	Year          string    `json:"Year"`
	ReleaseDate   string    `json:"ReleaseDate"`
	Studio        string    `json:"Studio"`
	MovieSet      string    `json:"MovieSet"`
	Plot          string    `json:"Plot"`
	Genres        []*Genre  `json:"Genres"`
	Actors        []*Actor  `json:"Actors"`
	Image         string    `json:"Image"`
	Thumb         string    `json:"Thumb"`
	Sample        []*Sample `json:"Sample"`
}

// Actor ...
type Actor struct {
	Image    string   `json:"Image"`
	StarLink string   `json:"StarLink"`
	Name     string   `json:"Name"`
	Alias    []string `json:"Alias"`
}

// Genre ...
type Genre struct {
	URL     string `json:"URL"`
	Content string `json:"Content"`
}

// Sample ...
type Sample struct {
	Index int64  `json:"Index"`
	Thumb string `json:"Thumb"`
	Image string `json:"Image"`
	Title string `json:"Title"`
}
