package conversion

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-cacher/cacher"
)

// Extend ...
type Extend struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// VideoSource ...
type VideoSource struct {
	Bangumi      string    `json:"bangumi"`       //番号 no
	VideoPath    string    `json:"video_path"`    //视频地址
	SourceHash   string    `json:"source_hash"`   //原片hash
	Type         string    `json:"type"`          //类型：film，FanDrama
	Format       string    `json:"format"`        //输出：3D，2D
	VR           string    `json:"vr"`            //VR格式：左右，上下，平面
	Thumb        string    `json:"thumb"`         //缩略图
	Intro        string    `json:"intro"`         //简介 title
	Alias        []string  `json:"alias"`         //别名，片名
	VideoEncode  string    `json:"video_encode"`  //视频编码
	AudioEncode  string    `json:"audio_encode"`  //音频编码
	Files        []string  `json:"files"`         //存放路径
	HashFiles    []string  `json:"hash_files"`    //已上传Hash
	CheckFiles   []string  `json:"check_files"`   //Unfinished checksum
	Slice        bool      `json:"sliceAdd"`      //是否HLS切片
	Encrypt      bool      `json:"encrypt"`       //加密
	Key          string    `json:"key"`           //秘钥
	M3U8         string    `json:"m3u8"`          //M3U8名
	SegmentFile  string    `json:"segment_file"`  //ts切片名
	PosterPath   string    `json:"poster_path"`   //海报路径
	Poster       string    `json:"poster"`        //海报HASH
	ExtendList   []*Extend `json:"extend_list"`   //扩展信息
	Role         []string  `json:"role"`          //角色列表 stars
	Director     string    `json:"director"`      //导演
	Systematics  string    `json:"systematics"`   //分级
	Season       string    `json:"season"`        //季
	Episode      string    `json:"episode"`       //集数
	TotalEpisode string    `json:"total_episode"` //总集数
	Sharpness    string    `json:"sharpness"`     //清晰度
	Publish      string    `json:"publish"`       //发行日
	Date         string    `json:"date"`          //发行日
	Length       string    `json:"length"`        //片长
	Producer     string    `json:"producer"`      //制片商
	Series       string    `json:"series"`        //系列
	Tags         []string  `json:"tags"`          //标签
	Publisher    string    `json:"publisher"`     //发行商
	Language     string    `json:"language"`      //语言
	Caption      string    `json:"caption"`       //字幕
	MagnetLinks  []string  `json:"magnet_links"`  //磁链
	Uncensored   bool      `json:"uncensored"`    //有码,无码
}

type sourceWalk struct {
	walk   *Walk
	source *VideoSource
}

// Walk ...
func (s *sourceWalk) Walk() Walk {
	return s.walk
}

// LoadWalk ...
func (s *sourceWalk) LoadWalk() error {
	bytes, e := cacher.Get(s.ID())
	if e != nil {
		return e
	}
	return json.Unmarshal(bytes, s.walk)
}

// ID ...
func (s sourceWalk) ID() string {
	return s.source.Bangumi
}

// Store ...
func (s *sourceWalk) Store() error {
	bytes, e := json.Marshal(s.walk)
	if e != nil {
		return e
	}
	return cacher.Set(s.ID(), bytes)
}

// Run ...
func (s *sourceWalk) Run(ctx context.Context) (e error) {
	v := s.source.Video()
	i, e := InsertOrUpdate(v)
	if e != nil {
		return e
	}
	if i == 0 {
		log.e
	}
}

// SourceWalk ...
func SourceWalk(source *VideoSource) IWalk {
	return &sourceWalk{
		walk: &Walk{
			ID:     source.Bangumi,
			Status: WalkWaiting,
			Value:  source,
		},
		source: source,
	}
}

// Video ...
func (v VideoSource) Video() *Video {
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
		Model:        Model{},
		BanNo:        strings.ToUpper(v.Bangumi),
		Intro:        intro,
		Alias:        v.Alias,
		ThumbHash:    "",
		PosterHash:   "",
		SourceHash:   "",
		M3U8Hash:     "",
		Key:          "",
		M3U8:         "",
		Role:         v.Role,
		Director:     v.Director,
		Systematics:  v.Systematics,
		Season:       MustString(v.Season, "1"),
		Episode:      MustString(v.Episode, "1"),
		TotalEpisode: MustString(v.TotalEpisode, "1"),
		Format:       MustString(v.Format, "2D"),
		Producer:     v.Producer,
		Publisher:    v.Publisher,
		Type:         v.Type,
		Language:     v.Language,
		Caption:      v.Caption,
		Group:        "",
		Index:        "",
		Date:         v.Date,
		Sharpness:    v.Sharpness,
		Series:       v.Series,
		Tags:         v.Tags,
		Length:       v.Length,
		Sample:       nil,
		Uncensored:   false,
	}
}
