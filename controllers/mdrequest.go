package controllers

import (
	"advapi/models"
	"advapi/tools"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"strings"
	"time"
)

// Operations about object
type MdRequestController struct {
	beego.Controller
}

func (this *MdRequestController) Get() {
	deviceName := this.Input().Get("ds")
	deviceMovement := this.Input().Get("dm")
	//screen := this.Input().Get("screen")
	placementHash := this.Input().Get("s")
	advType := this.Input().Get("mt")
	screen := this.Input().Get("screen")
	equipmentKey := this.Input().Get("i")

	zone := models.GetZone(placementHash)

	if zone == nil {
		this.Data["ResCode"] = models.ResCodeInputError
		this.TplNames = "error.tpl"
		return
	}
	clientIp := GetClientIP(this.Ctx.Input)

	aryPackageId := models.GetDeviceQuality(models.GetDevice(models.Device{DeviceName: deviceName, DeviceMovment: deviceMovement}))

	//广西：116.252.146.109
	//上海：124.77.127.23
	//山东：27.193.0.0
	provinceCode, cityCode := models.GetLocationCodes("116.252.146.109")

	adUnit := models.LaunchQuery(provinceCode, cityCode, zone, aryPackageId, advType, screen)

	requestLog := tools.DeviceRequestLog{
		EquipmentKey:  equipmentKey,
		DeviceName:    deviceName,
		Date:          time.Now().Format("2006-01-02 15:04:05"),
		PublicationId: zone.PublicationId,
		ZoneId:        zone.EntryId,
		ClientIp:      clientIp,
		ProvinceCode:  provinceCode,
		CityCode:      cityCode,
		BusinessId:    "MDADV",
	}

	operationType := "001"
	if adUnit != nil {
		operationType = "002"
		requestLog.CampaignId = adUnit.CampaignId
		requestLog.CreativeId = adUnit.AdvId
	}

	requestLog.OperationType = operationType

	tools.SendDeviceRequestLog(&requestLog)

	beego.Debug(screen, advType, zone, clientIp, aryPackageId, provinceCode, cityCode)

	this.TplNames = "advresponse.tpl"
	//this.ServeJson()
}

func GetClientIP(input *context.BeegoInput) string {
	ips := input.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	ip := strings.Split(input.Request.RemoteAddr, ":")
	if len(ip) > 0 {
		return ip[0]
	}
	return "127.0.0.1"
}
