package conversion

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// Video ...
type Video struct {
	Model `xorm:"extends" json:"-"`
	//FindNo       string    `xorm:"find_no" json:"-"`               //查找号
	BanNo        string   `xorm:"ban_no" json:"ban_no"`               //番号
	Intro        string   `xorm:"varchar(2048)" json:"intro"`         //简介
	Alias        []string `xorm:"json" json:"alias"`                  //别名，片名
	ThumbHash    string   `xorm:"thumb_hash" json:"thumb_hash"`       //缩略图
	PosterHash   string   `xorm:"poster_hash" json:"poster_hash"`     //海报地址
	SourceHash   string   `xorm:"source_hash" json:"source_hash"`     //原片地址
	M3U8Hash     string   `xorm:"m3u8_hash" json:"m3u8_hash"`         //切片地址
	Key          string   `xorm:"key"  json:"-"`                      //秘钥
	M3U8         string   `xorm:"m3u8" json:"-"`                      //M3U8名
	Role         []string `xorm:"json" json:"role"`                   //主演
	Director     string   `xorm:"director" json:"director"`           //导演
	Systematics  string   `xorm:"systematics" json:"systematics"`     //分级
	Season       string   `xorm:"season" json:"season"`               //季
	TotalEpisode string   `xorm:"total_episode" json:"total_episode"` //总集数
	Episode      string   `xorm:"episode" json:"episode"`             //集数
	Producer     string   `xorm:"producer" json:"producer"`           //生产商
	Publisher    string   `xorm:"publisher" json:"publisher"`         //发行商
	Type         string   `xorm:"type" json:"type"`                   //类型：film，FanDrama
	Format       string   `xorm:"format" json:"format"`               //输出格式：3D，2D,VR(VR格式：Half-SBS：左右半宽,Half-OU：上下半高,SBS：左右全宽)
	Language     string   `xorm:"language" json:"language"`           //语言
	Caption      string   `xorm:"caption" json:"caption"`             //字幕
	Group        string   `xorm:"group" json:"-"`                     //分组
	Index        string   `xorm:"index" json:"-"`                     //索引
	Date         string   `xorm:"'date'" json:"date"`                 //发行日期
	Sharpness    string   `xorm:"sharpness" json:"sharpness"`         //清晰度
	Series       string   `xorm:"series" json:"series"`               //系列
	Tags         []string `xorm:"json tags" json:"tags"`              //标签
	Length       string   `xorm:"length" json:"length"`               //时长
	Sample       []string `xorm:"json sample" json:"sample"`          //样板图
	Uncensored   bool     `xorm:"uncensored" json:"uncensored"`       //有码,无码
	//MagnetLinks  []string  `json:"-"`                              //磁链
	//Visit        uint64    `json:"-" xorm:"notnull default(0)"`    //访问数
}

func init() {
	registerTable(&Video{})
}

// BeforeInsert ...
func (v *Video) BeforeInsert() {
	v.Model.BeforeInsert()
	if len(v.Role) == 0 {
		v.Role = []string{}
	}
	if len(v.Alias) == 0 {
		v.Alias = []string{}
	}
	if len(v.Tags) == 0 {
		v.Tags = []string{}
	}
	if len(v.Sample) == 0 {
		v.Sample = []string{}
	}
}

// Sync ...
func (v *Video) Sync() error {
	return _database.Sync2(v)
}

// Table ...
func (v *Video) Table() interface{} {
	return &Video{}
}

// IVideo ...
type IVideo interface {
	Video() *Video
}

// Checksum ...
func Checksum(filepath string) string {
	hash := sha256.New()
	file, e := os.Open(filepath)
	if e != nil {
		return ""
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	_, e = io.Copy(hash, reader)
	if e != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
