package conversion

import (
	"context"
	"encoding/json"
)

// VideoPornInfo ...
type VideoPornInfo struct {
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
func (v VideoPornInfo) Video() *Video {
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
		No:           v.ID,
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

// NewInfoWork ...
func NewInfoWork(info *VideoPornInfo, options ...WorkOptions) (IWork, error) {
	bys, e := json.Marshal(info)
	if e != nil {
		return nil, e
	}

	options = append(options, IDOption(info.ID))
	work := newWork("info", defaultWork(options...), bys)

	if work.ID() == "" {
		return nil, ErrWorkID
	}

	return work, nil
}
func decodeInfo(src []byte) (IVideo, error) {
	var info VideoPornInfo
	e := json.Unmarshal(src, &info)
	if e != nil {
		return nil, e
	}
	return &info, nil
}

// VideoFromInfo ...
func VideoFromInfo(ctx context.Context, Work *Work) (IVideo, error) {
	log.Info("info process run")
	info, e := decodeInfo(Work.Value)
	if e != nil {
		return nil, e
	}

	return info, nil
}
