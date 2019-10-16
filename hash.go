package conversion

import (
	"errors"

	"github.com/go-xorm/xorm"
)

// Type ...
type Type string

// TypeOther ...
const TypeOther Type = "other"

// TypeVideo ...
const TypeVideo Type = "video"

// TypeSlice ...
const TypeSlice Type = "slice"

// TypePoster ...
const TypePoster Type = "poster"

// TypeThumb ...
const TypeThumb Type = "thumb"

// TypeCaption caption file
const TypeCaption Type = "caption"

// Hash ...
type Hash struct {
	Model       `xorm:"extends"`
	Checksum    string `xorm:"default() checksum"`            //sum值
	Type        Type   `xorm:"default() type"`                //类型
	Relate      string `xorm:"default()" json:"relate"`       //关联信息
	Name        string `xorm:"default() name"`                //名称
	Hash        string `xorm:"default() hash"`                //哈希地址
	Sharpness   string `xorm:"default()" json:"sharpness"`    //清晰度
	Caption     string `xorm:"default()" json:"caption"`      //字幕
	Encrypt     bool   `json:"encrypt"`                       //加密
	Key         string `xorm:"default()" json:"key"`          //秘钥
	M3U8        string `xorm:"m3u8 default()" json:"m3u8"`    //M3U8名
	SegmentFile string `xorm:"default()" json:"segment_file"` //ts切片名
	Resource    string `xorm:"notnull default(0)"`            //资源地址
}

// Table ...
func (h *Hash) Table() interface{} {
	return &Hash{}
}

func init() {
	registerTable(&Hash{})
}

// Sync ...
func (h *Hash) Sync() error {
	return _database.Sync2(h)
}

// AllHash ...
func AllHash(session *xorm.Session, limit int, start ...int) (unfins *[]*Hash, e error) {
	unfins = new([]*Hash)
	session = MustSession(session)
	if limit > 0 {
		session = session.Limit(limit, start...)
	}
	if err := session.Find(unfins); err != nil {
		return nil, err
	}
	return unfins, nil
}

// FindHash ...
func FindHash(session *xorm.Session, checksum string) (unfin *Hash, e error) {
	unfin = new(Hash)
	b, e := MustSession(session).Where("checksum = ?", checksum).Get(unfin)
	if e != nil || !b {
		return nil, errors.New("hash not found")
	}
	return unfin, nil
}

// AddOrUpdateHash ...
func AddOrUpdateHash(session *xorm.Session, unfin *Hash) (e error) {
	tmp := new(Hash)
	var found bool
	session = MustSession(session)
	if unfin.ID() != "" {
		found, e = session.Clone().ID(unfin.ID).Get(tmp)
	} else {
		found, e = session.Clone().Where("checksum = ?", unfin.Checksum).
			Where("type = ?", unfin.Type).Get(tmp)
	}
	if e != nil {
		return e
	}
	if found {
		//only slice need update,video update for check , hash changed
		i := int64(0)
		if unfin.Hash != unfin.Hash || unfin.Type == TypeSlice || unfin.Type == TypeVideo {
			unfin.SetVersion(tmp.Version())
			unfin.SetID(tmp.ID())
			i, e = session.Clone().ID(unfin.ID).Update(unfin)
			log.Infof("updated(%d): %+v", i, tmp)
		}
		return e
	}
	_, e = session.Clone().InsertOne(unfin)
	return
}

// Clone ...
func (h *Hash) Clone() (n *Hash) {
	n = new(Hash)
	*n = *h
	return
}
