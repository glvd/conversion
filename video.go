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
	BanNo        string    `xorm:"ban_no" json:"ban_no"`           //番号
	Intro        string    `xorm:"varchar(2048)" json:"intro"`     //简介
	Alias        []string  `xorm:"json" json:"alias"`              //别名，片名
	ThumbHash    string    `xorm:"thumb_hash" json:"thumb_hash"`   //缩略图
	PosterHash   string    `xorm:"poster_hash" json:"poster_hash"` //海报地址
	SourceHash   string    `xorm:"source_hash" json:"source_hash"` //原片地址
	M3U8Hash     string    `xorm:"m3u8_hash" json:"m3u8_hash"`     //切片地址
	Key          string    `xorm:"key"  json:"-"`                  //秘钥
	M3U8         string    `xorm:"m3u8" json:"-"`                  //M3U8名
	Role         []string  `xorm:"json" json:"role"`               //主演
	Director     string    `xorm:"director" json:"director"`       //导演
	Systematics  string    `json:"systematics"`                    //分级
	Season       string    `json:"season"`                         //季
	TotalEpisode string    `json:"total_episode"`                  //总集数
	Episode      string    `json:"episode"`                        //集数
	Producer     string    `json:"producer"`                       //生产商
	Publisher    string    `json:"publisher"`                      //发行商
	Type         string    `json:"type"`                           //类型：film，FanDrama
	Format       string    `json:"format"`                         //输出格式：3D，2D,VR(VR格式：Half-SBS：左右半宽,Half-OU：上下半高,SBS：左右全宽)
	Language     string    `json:"language"`                       //语言
	Caption      string    `json:"caption"`                        //字幕
	Group        string    `json:"-"`                              //分组
	Index        string    `json:"-"`                              //索引
	Date         string    `json:"date"`                           //发行日期
	Sharpness    string    `json:"sharpness"`                      //清晰度
	Series       string    `json:"series"`                         //系列
	Tags         []string  `xorm:"tags" json:"tags"`               //标签
	Length       string    `json:"length"`                         //时长
	Sample       []*Sample `json:"sample"`                         //样板图
	Uncensored   bool      `json:"uncensored"`                     //有码,无码
	//MagnetLinks  []string  `json:"-"`                              //磁链
	//Visit        uint64    `json:"-" xorm:"notnull default(0)"`    //访问数
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
