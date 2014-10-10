package tools

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/kr/beanstalk"
)

type DeviceRequestLog struct {
	EntryId        int
	EquipmentSn    string
	EquipmentKey   string
	DeviceId       int
	DeviceName     string
	UserPattern    string
	Date           string
	OperationType  string
	OperationExtra string
	PublicationId  int
	ZoneId         int
	CampaignId     int
	CreativeId     int
	ClientIp       string
	ProvinceCode   string
	CityCode       string
	BusinessId     string
}

var reqChan chan *DeviceRequestLog
var requestTube beanstalk.Tube

func InitRequestPool(address, tubeName string) {
	reqChan = make(chan *DeviceRequestLog, 1000)
	go processDeviceRequest(address, tubeName)
}

func SendDeviceRequestLog(deviceRequestLog *DeviceRequestLog) {
	reqChan <- deviceRequestLog
}

//func SendDeviceRequestLog(reqChan *DeviceRequestLog) error {

//	strLog, _ := json.Marshal(deviceRequestLog)
//	beego.Debug(string(strLog))

//	c, _ := beanstalk.Dial("tcp", "127.0.0.1:11300")
//	tube := beanstalk.Tube{c, "MDADV_REQUEST_DEVICE_LOG"}

//	_, err := tube.Put([]byte(strLog), 0, 0, 0)

//	beego.Debug(err)
//	return err
//}

func processDeviceRequest(address, tubeName string) {
	c, error := beanstalk.Dial("tcp", address)
	if error != nil {
		beego.Error("Can not connect to Beanstalkd server!")
		return
	}
	requestTube := beanstalk.Tube{c, tubeName}
	defer c.Close()

	for {

		deviceRequestLog := <-reqChan
		strLog, error := json.Marshal(deviceRequestLog)
		if error != nil {
			beego.Error(error)
		}
		beego.Debug(string(strLog))

		requestTube.Put([]byte(strLog), 0, 0, 0)

	}
}
