package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"net/http"
	"time"
)

// 向第三方发送异常报警
func SendToThirdPartyAPI(exception *Exception) {
	exception.Timestamp = time.Now().Format("20060102150405")
	exception.ServerIP = RecordPluginConfig.LocalIp
	data, err := json.Marshal(exception)
	if err != nil {
		fmt.Println("Error marshalling exception:", err)
		return
	}

	resp, err := http.Post(RecordPluginConfig.ExceptionPostUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error sending exception to third party API:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to send exception, status code:", resp.StatusCode)
	} else {
		fmt.Println("Exception sent successfully!")
	}
}

// 磁盘超上限报警
func getDisckException(streamPath string) bool {
	d, _ := disk.Usage("/")
	if d.UsedPercent >= RecordPluginConfig.DiskMaxPercent {
		exceptionChannel <- &Exception{AlarmType: "disk alarm", AlarmDesc: "disk is full", Channel: streamPath}
		return true
	}
	return false
}
