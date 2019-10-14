package conversion

// Video ...
type Video struct {
	Model        `xorm:"extends" json:"-"`
	FindNo       string   `json:"-"`                              //查找号
	Bangumi      string   `xorm:"bangumi" json:"bangumi"`         //番組
	Intro        string   `xorm:"varchar(2048)" json:"intro"`     //简介
	Alias        []string `xorm:"json" json:"alias"`              //别名，片名
	ThumbHash    string   `xorm:"thumb_hash" json:"thumb_hash"`   //缩略图
	PosterHash   string   `xorm:"poster_hash" json:"poster_hash"` //海报地址
	SourceHash   string   `xorm:"source_hash" json:"source_hash"` //原片地址
	M3U8Hash     string   `xorm:"m3u8_hash" json:"m3u8_hash"`     //切片地址
	Key          string   `json:"-"`                              //秘钥
	M3U8         string   `xorm:"m3u8" json:"-"`                  //M3U8名
	Role         []string `xorm:"json" json:"role"`               //主演
	Director     string   `json:"-"`                              //导演
	Systematics  string   `json:"-"`                              //分级
	Season       string   `json:"-"`                              //季
	TotalEpisode string   `json:"-"`                              //总集数
	Episode      string   `json:"-"`                              //集数
	Producer     string   `json:"-"`                              //生产商
	Publisher    string   `json:"-"`                              //发行商
	Type         string   `json:"-"`                              //类型：film，FanDrama
	Format       string   `json:"format"`                         //输出格式：3D，2D,VR(VR格式：Half-SBS：左右半宽,Half-OU：上下半高,SBS：左右全宽)
	Language     string   `json:"-"`                              //语言
	Caption      string   `json:"-"`                              //字幕
	Group        string   `json:"-"`                              //分组
	Index        string   `json:"-"`                              //索引
	Date         string   `json:"-"`                              //发行日期
	Sharpness    string   `json:"sharpness"`                      //清晰度
	Visit        uint64   `json:"-" xorm:"notnull default(0)"`    //访问数
	Series       string   `json:"series"`                         //系列
	Tags         []string `xorm:"json" json:"tags"`               //标签
	Length       string   `json:"length"`                         //时长
	MagnetLinks  []string `json:"-"`                              //磁链
	Uncensored   bool     `json:"uncensored"`                     //有码,无码
}

// IVideo ...
type IVideo interface {
	Video() *Video
}
