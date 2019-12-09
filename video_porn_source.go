package conversion

import (
	"encoding/json"
	"strings"
)

// Extend ...
type Extend struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// VideoPornSource ...
type VideoPornSource struct {
	Bangumi    string    `json:"bangumi"`     //番号 no
	VideoPath  []string  `json:"video_path"`  //视频地址
	SourceHash string    `json:"source_hash"` //原片hash
	ThumbPath  string    `json:"thumb_path"`  //缩略图路径
	PosterPath string    `json:"poster_path"` //海报路径
	Format     string    `json:"format"`      //输出：3D，2D
	Thumb      string    `json:"thumb"`       //缩略图HASH
	Poster     string    `json:"poster"`      //海报HASH
	Intro      string    `json:"intro"`       //简介 title
	Alias      []string  `json:"alias"`       //别名，片名
	Role       []string  `json:"role"`        //角色列表 stars
	Director   string    `json:"director"`    //导演
	Date       string    `json:"date"`        //发行日
	Length     string    `json:"length"`      //片长
	Producer   string    `json:"producer"`    //制片商
	Series     string    `json:"series"`      //系列
	Tags       []string  `json:"tags"`        //标签
	Publisher  string    `json:"publisher"`   //发行商
	Language   string    `json:"language"`    //语言
	Caption    string    `json:"caption"`     //字幕
	Uncensored bool      `json:"uncensored"`  //有码,无码
	ExtendList []*Extend `json:"extend_list"` //扩展信息
}

// NewSourceWork ...
func NewSourceWork(source *VideoPornSource, options ...WorkOptions) (IWork, error) {
	bys, e := json.Marshal(source)
	if e != nil {
		return nil, e
	}
	opts := []WorkOptions{IDOption(source.Bangumi),
		VideoPathOption(source.VideoPath),
		PosterPathOption(source.PosterPath),
		ThumbPathOption(source.Thumb)}
	opts = append(opts, options...)
	work := newWork("source", defaultWork(options...), bys)
	return work, nil
}

func decodeSource(src []byte) (IVideo, error) {
	var source VideoPornSource
	e := json.Unmarshal(src, &source)
	if e != nil {
		return nil, e
	}
	return &source, nil
}

// VideoFromSource ...
func VideoFromSource(Work *Work) (IVideo, error) {
	log.Info("source process run")
	source, e := decodeSource(Work.Value)
	if e != nil {
		return nil, e
	}
	//v := source.Video()
	//i, e := InsertOrUpdate(v)
	//if e != nil {
	//	return e
	//}
	//if i == 0 {
	//	log.With("id", source.Bangumi).Warn("not updated")
	//}
	return source, nil
}

// Video ...
func (v VideoPornSource) Video() *Video {
	//always not null
	alias := ""
	if len(v.Alias) > 0 {
		alias = v.Alias[0]
	}
	//always not null
	role := ""
	if len(v.Role) > 0 {
		role = v.Role[0]
	}

	intro := v.Intro
	if intro == "" {
		intro = alias + " " + role
	}

	return &Video{
		Model:      Model{},
		No:         strings.ToUpper(v.Bangumi),
		Intro:      intro,
		Alias:      v.Alias,
		ThumbHash:  "",
		PosterHash: "",
		SourceHash: "",
		M3U8Hash:   "",
		Key:        "",
		M3U8:       "",
		Role:       v.Role,
		Director:   v.Director,
		////Systematics:  v.Systematics,
		//Season:       MustString(v.Season, "1"),
		//Episode:      MustString(v.Episode, "1"),
		//TotalEpisode: MustString(v.TotalEpisode, "1"),
		Format:    MustString(v.Format, "2D"),
		Producer:  v.Producer,
		Publisher: v.Publisher,
		//Type:         v.Type,
		Language: v.Language,
		Caption:  v.Caption,
		Group:    "",
		Index:    "",
		Date:     v.Date,
		//Sharpness:    v.Sharpness,
		Series:     v.Series,
		Tags:       v.Tags,
		Length:     v.Length,
		Sample:     nil,
		Uncensored: false,
	}
}
