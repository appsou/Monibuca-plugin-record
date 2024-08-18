package record

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"

	"m7s.live/engine/v4/util"
)

func errorJsonString(args map[string]interface{}) string {
	resultJsonData := make(map[string]interface{})
	for field, value := range args {
		resultJsonData[field] = value
	}
	jsonString, _ := json.Marshal(resultJsonData)
	return string(jsonString)
}

// 事件录像
func (conf *RecordConfig) API_event_list(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 用token验证用户信息是否有效，并获取用户信息换取userid
	var postData map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	err = json.Unmarshal(body, &postData)
	pageNum := postData["pageNum"].(float64)
	if pageNum <= 0 {
		resultJsonData["msg"] = "pageNum error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	pageSize := postData["pageSize"].(float64)
	if pageSize <= 0 {
		resultJsonData["msg"] = "pageSize error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	eventRecords, totalCount, err := paginate(EventRecord{}, int(pageNum), int(pageSize), postData)
	if err != nil {
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	resultJsonData["totalCount"] = totalCount
	resultJsonData["pageNum"] = pageNum
	resultJsonData["pageSize"] = pageSize
	resultJsonData["list"] = eventRecords
	resultJsonData["code"] = 0
	resultJsonData["msg"] = ""
	util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
}

// 事件录像
func (conf *RecordConfig) API_event_start(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 用token验证用户信息是否有效，并获取用户信息换取userid
	var eventRecordMode EventRecord
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	// 解析JSON数据到map
	err = json.Unmarshal(body, &eventRecordMode)
	if err != nil {
		resultJsonData["msg"] = "Invalid JSON format "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	streamPath := eventRecordMode.StreamPath
	if streamPath == "" {
		resultJsonData["msg"] = "no streamPath"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 获取到磁盘容量低，磁盘报错的情况下需要报异常，并且根据事件类型做出处理
	if getDisckException(streamPath) {
		resultJsonData["msg"] = "disk is full"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//var streamExist = false
	//conf.recordings.Range(func(key, value any) bool {
	//	existStreamPath := value.(IRecorder).GetSubscriber().Stream.Path
	//	if existStreamPath == streamPath {
	//		resultJsonData["msg"] = "streamPath is exist"
	//		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
	//		streamExist = true
	//		return !streamExist
	//	}
	//	return !streamExist
	//})
	//if streamExist {
	//	return
	//}
	eventId := eventRecordMode.EventId
	if eventId == "" {
		resultJsonData["msg"] = "no eventId"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	eventType := eventRecordMode.EventType
	if eventType == "" {
		resultJsonData["msg"] = "no eventType"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	eventName := eventRecordMode.EventName
	if eventName == "" {
		resultJsonData["msg"] = "no eventName"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	beforeDuration := eventRecordMode.BeforeDuration
	if beforeDuration == "" {
		beforeDuration = strconv.Itoa(conf.beforeDuration)
	}
	afterDuration := eventRecordMode.AfterDuration
	if afterDuration == "" {
		afterDuration = strconv.Itoa(conf.afterDuration)
	}
	recordTime := time.Now().Format("2006-01-02 15:04:05")
	fileName := time.Now().Format("20060102150405")
	startTime := time.Now().Add(-30 * time.Second).Format("2006-01-02 15:04:05")
	endTime := time.Now().Add(30 * time.Second).Format("2006-01-02 15:04:05")
	//切片大小
	fragment := eventRecordMode.Fragment
	//var id string
	irecorder := NewFLVRecorder()
	recorder := irecorder.GetRecorder()
	recorder.FileName = fileName
	recorder.append = false
	filepath := conf.Flv.Path + "/" + streamPath + "/" + fileName + recorder.Ext
	if fragment != "" {
		if f, err := time.ParseDuration(fragment); err == nil {
			recorder.Fragment = f
		}
	}
	err = irecorder.StartWithFileName(streamPath, fileName)
	go func() {
		timer := time.NewTimer(30 * time.Second)

		// 等待计时器到期
		<-timer.C
		recorder.Stop(zap.String("reason", "api"))
	}()
	//id = recorder.ID
	if err != nil {
		exceptionChannel <- &Exception{AlarmType: "record", AlarmDesc: "录像失败", Channel: streamPath}
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	eventRecordMode.Filepath = filepath
	eventRecordMode.Filename = fileName + recorder.Ext
	eventRecord := EventRecord{StreamPath: streamPath, EventId: eventId, EventType: eventType, EventName: eventName, BeforeDuration: beforeDuration,
		AfterDuration: afterDuration, RecordTime: recordTime, StartTime: startTime, EndTime: endTime, Filepath: filepath, Filename: fileName + recorder.Ext}
	err = mysqldb.Omit("id", "fragment", "isDelete").Create(&eventRecord).Error
	if err != nil {
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	outid := eventRecord.Id
	resultJsonData["id"] = outid
	resultJsonData["code"] = 0
	resultJsonData["msg"] = ""
	util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
}
