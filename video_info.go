package conversion

import "encoding/json"

// VideoInfo ...
type VideoInfo struct {
	From          string    `json:"From"`          //来源
	Uncensored    bool      `json:"Uncensored"`    //无码
	ID            string    `json:"ID"`            //番号
	Title         string    `json:"Title"`         //标题
	OriginalTitle string    `json:"OriginalTitle"` //原始标题
	Year          string    `json:"Year"`          //日期
	ReleaseDate   string    `json:"ReleaseDate"`   //发行日
	Studio        string    `json:"Studio"`        //工作室
	MovieSet      string    `json:"MovieSet"`      //系列
	Plot          string    `json:"Plot"`          //情节
	Genres        []*Genre  `json:"Genres"`        //类型,标签
	Actors        []*Actor  `json:"Actors"`        //演员
	Image         string    `json:"Image"`         //海报
	Thumb         string    `json:"Thumb"`         //缩略图
	Sample        []*Sample `json:"Sample"`        //样板图
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

// Video ...
func (v VideoInfo) Video() *Video {
	var role []string
	for _, act := range v.Actors {
		role = append(role, act.Name)
	}
	var tags []string
	for _, gen := range v.Genres {
		tags = append(tags, gen.Content)
	}

	return &Video{
		Model:        Model{},
		BanNo:        v.ID,
		Intro:        v.Title,
		Alias:        nil,
		ThumbHash:    "",
		PosterHash:   "",
		SourceHash:   "",
		M3U8Hash:     "",
		Key:          "",
		M3U8:         "",
		Role:         role,
		Director:     "",
		Systematics:  "",
		Season:       MustString("", "1"),
		TotalEpisode: MustString("", "1"),
		Episode:      MustString("", "1"),
		Producer:     v.Studio,
		Publisher:    "",
		Type:         "",
		Format:       MustString("", "2D"),
		Language:     "",
		Caption:      "",
		Group:        "",
		Index:        "",
		Date:         v.ReleaseDate,
		Sharpness:    "",
		Series:       v.MovieSet,
		Tags:         tags,
		Length:       "",
		Sample:       nil,
		Uncensored:   v.Uncensored,
	}
}

// NewInfoWalk ...
func NewInfoWalk(info *VideoInfo, options ...WalkOptions) (IWalk, error) {
	bytes, e := json.Marshal(info)
	if e != nil {
		return nil, e
	}
	walk := &Walk{
		WalkImpl: WalkImpl{
			ID:       info.ID,
			WalkType: "info",
			Status:   WalkWaiting,
			Value:    bytes,
		},
	}
	for _, opt := range options {
		opt(walk)
	}
	return walk, nil
}
func decodeInfo(src []byte) (*VideoInfo, error) {
	var info VideoInfo
	e := json.Unmarshal(src, &info)
	if e != nil {
		return nil, e
	}
	return &info, nil
}

// InfoProcess ...
func InfoProcess(walk *Walk) error {
	log.Info("info process run")
	info, e := decodeInfo(walk.Value)
	if e != nil {
		return e
	}
	v := info.Video()
	i, e := InsertOrUpdate(v)
	if e != nil {
		return e
	}
	if i == 0 {
		log.With("id", info.ID).Warn("not updated")
	}
	return nil
}
