package conversion

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
