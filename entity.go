package record

import "time"

// mysql数据库eventrecord表
type EventRecord struct {
	Id             uint   `json:"id" desc:"自增长id" gorm:"primaryKey"`
	StreamPath     string `json:"streamPath" desc:"流路径" gorm:"column:streamPath"`
	EventId        string `json:"eventId" desc:"事件编号" gorm:"column:eventId"`
	EventType      string `json:"eventType" desc:"事件类型" gorm:"column:eventType"`
	EventName      string `json:"eventName" desc:"事件名称" gorm:"column:eventName"`
	BeforeDuration string `json:"beforeDuration" desc:"事件前缓存时长" gorm:"column:beforeDuration"`
	AfterDuration  string `json:"afterDuration" desc:"事件后缓存时长" gorm:"column:afterDuration"`
	RecordTime     string `json:"recordTime" desc:"录像时间" gorm:"column:recordTime"`
	StartTime      string `json:"startTime" desc:"录像开始时间" gorm:"column:startTime"`
	EndTime        string `json:"endTime" desc:"录像结束时间" gorm:"column:endTime"`
	Filepath       string `json:"filepath" desc:"录像文件路径" gorm:"column:filepath"`
	IsDelete       string `json:"isDelete" desc:"是否删除，0表示正常，1表示删除，默认0" gorm:"column:isDelete"`
	UserId         string `json:"useId" desc:"用户id" gorm:"-;column:useId"`
	Filename       string `json:"filename" desc:"文件名" gorm:"column:filename"`
	Fragment       string `json:"fragment" desc:"切片大小" gorm:"-"`
}

// TableName 返回自定义的表名
func (EventRecord) TableName() string {
	return "eventrecord"
}

// mysql数据库里Exception 定义异常结构体
type Exception struct {
	Timestamp string `json:"Timestamp" gorm:"autoCreateTime"`
	AlarmType string `json:"AlarmType"`
	AlarmDesc string `json:"AlarmDesc"`
	ServerIP  string `json:"ServerIP"`
	Channel   string `json:"Channel"`
}

// sqlite数据库用来存放每个flv文件的关键帧对应的offset及abstime数据
type FLVKeyframe struct {
	FLVFileName  string    `gorm:"not null"`
	FrameOffset  int64     `gorm:"not null"`
	FrameAbstime uint32    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
