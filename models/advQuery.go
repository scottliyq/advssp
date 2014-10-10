package models

import (
	"advssp/tools"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"strings"
	"sync"
	"time"
)

const ResCodeSuccess = "00000"
const ResCodeInputError = "30001"
const ResCodeNoAd = "20001"

var once sync.Once
var cities (map[string]string)
var regions (map[string]string)

func init() {

}

func GetZone(zoneHash string) (rstZone *Zone) {
	o := orm.NewOrm()

	params := []string{zoneHash}

	err := o.Raw("SELECT entry_id,publication_id,zone_width,zone_height,zone_lastrequest FROM md_zones WHERE zone_hash = ?", params).QueryRow(&rstZone)

	if err == nil && &rstZone != nil {
		beego.Debug("Check if zone hash is available: ", true)
		return rstZone
	}

	return nil
}

func GetDevice(device Device) (deviceId int) {
	o := orm.NewOrm()

	var rstDevice Device
	var err error
	if device.DeviceName == "" {
		params := []string{device.DeviceMovment}
		err = o.Raw("SELECT entry_id FROM md_devices WHERE device_name = ?", params).QueryRow(&rstDevice)
	} else {
		params := []string{device.DeviceMovment, device.DeviceName}
		err = o.Raw("SELECT entry_id FROM md_devices WHERE device_name = ? and device_movement = ?", params).QueryRow(&rstDevice)
	}

	if err == nil && &rstDevice != nil {
		return rstDevice.DeviceId
	}

	return 0
}

func GetDeviceQuality(deviceId int) (ret []int) {

	if deviceId == 0 {
		return nil
	}
	o := orm.NewOrm()

	var list []DevicePackageMatrix
	num, err := o.Raw("SELECT package_id FROM user WHERE device_id = ?", deviceId).QueryRows(&list)
	if err == nil && num > 0 {
		beego.Debug(list) // []{"1","2","3",...}
		size := len(list)
		ret := make([]int, size)
		for i := 0; i < size; i++ {
			ret[i] = list[i].PackageId
		}
		return ret
	}

	return nil
}

func UpdateLastRequest(zone Zone) (err error) {

	o := orm.NewOrm()

	_, err = o.Raw("UPDATE md_zones SET zone_lastrequest = current_timestamp() WHERE entry_id", zone.EntryId).Exec()
	if err != nil {
		return err
	}
	_, err = o.Raw("UPDATE md_publications SET zone_lastrequest = current_timestamp() WHERE inv_id", zone.EntryId).Exec()
	if err != nil {
		return err
	}

	return nil
}

func initMap() {
	cities = map[string]string{
		"CN_01": "北京市",
		"CN_02": "天津市",
		"CN_09": "上海市",
		"CN_22": "重庆市",
		"CN_32": "香港",
		"CN_33": "澳门",
	}

	regions = map[string]string{
		"CN_05": "内蒙古",
		"CN_20": "广西",
		"CN_26": "西藏",
		"CN_30": "宁夏",
		"CN_31": "新疆",
	}

}

func GetLocationCodes(address string) (provinceCode, cityCode string) {

	once.Do(initMap)

	location := tools.QueryIP(address)
	beego.Debug(cities, regions, location)
	beego.Debug("Find address stridng:", location)

	checkFlg := true

	if location != "" {
		for k, v := range cities {
			if strings.Contains(location, v) {
				provinceCode, cityCode = k, k
				checkFlg = false
				break
			}
		}

		for k, v := range regions {
			if strings.Contains(location, v) {
				provinceCode = k
				cityName := strings.Replace(location, v, "", 1)
				beego.Debug(cityName)
				cityCode = getCodeFromDB(cityName)

				checkFlg = false
				break
			}
		}

		if checkFlg {
			if strings.Contains(location, "省") {
				aryTemp := strings.Split(location, "省")
				provinceCode = getCodeFromDB(aryTemp[0] + "省")
				cityCode = getCodeFromDB(aryTemp[1])
			}
		}

	}

	return provinceCode, cityCode
}

func getCodeFromDB(regionName string) string {
	o := orm.NewOrm()

	//params := []string{regionName}
	var rstRegion *RegionalTargeting
	err := o.Raw("SELECT targeting_code FROM md_regional_targeting WHERE region_name = ? limit 1", regionName).QueryRow(&rstRegion)

	if err == nil && &rstRegion != nil {
		return rstRegion.TargetingCode
	}

	return ""
}

func DeductImpressionNumber(campaignId, number int) (err error) {

	o := orm.NewOrm()

	_, err = o.Raw("UPDATE md_campaign_limit SET total_amount_left = total_amount_left - ? WHERE campaign_id = ? AND total_amount_left>0", number, campaignId).Exec()
	if err != nil {
		return err
	}

	return nil
}

func buildQuery(provinceCode string, cityCode string, zone *Zone, qualityIds []int, advType string, screenCode string) (strQuery string, leftGeo bool, leftQuality bool) {
	conditions := make([]string, 10, 10)
	i := 0
	conditions[i] = " (Campaigns.country_target=1"
	i = i + 1

	isProvinceTarget := isExistTargeting("geo", provinceCode)
	isCityTarget := isExistTargeting("geo", cityCode)

	if isProvinceTarget || isCityTarget {
		if provinceCode != "" && cityCode != "" {
			conditions[i] = fmt.Sprintf(" OR (c1.targeting_type='geo' AND (c1.targeting_code=%q OR c1.targeting_code=%q)))", provinceCode, cityCode)
			i = i + 1
		} else if provinceCode != "" {

			conditions[i] = fmt.Sprintf(" OR (c1.targeting_type='geo' AND c1.targeting_code=%q))", provinceCode)
			i = i + 1
		} else {
			conditions[i] = ")"
			i = i + 1
		}
		leftGeo = true
	} else {
		conditions[i] = "(Campaigns.country_target=1)"
		i = i + 1
		leftGeo = false
	}

	conditions[i] = fmt.Sprintf(" AND (c3.targeting_type='placement' AND c3.targeting_code=%d)", zone.EntryId)
	i = i + 1

	isQualityTarget := isExistQualityTargeting(qualityIds)

	if isQualityTarget {
		conditions[i] = fmt.Sprintf(" AND (Campaigns.quality_target=1 OR (c7.targeting_type='quality' AND c7.targeting_code IN (%q)", composeString(qualityIds))
		i = i + 1
		leftQuality = true
	} else {
		conditions[i] = " AND (Campaigns.quality_target=1)"
		i = i + 1
		leftQuality = false
	}

	currentDate := time.Now().Format("2006-01-02")

	//投放时间条件
	conditions[i] = fmt.Sprintf(" AND Campaigns.campaign_status=1 AND Campaigns.campaign_class<>2 AND Campaigns.campaign_start<=%q AND Campaigns.campaign_end>=%q", currentDate, currentDate)
	i = i + 1

	//创意类型条件

	if zone.ZoneType == "open" {

		if advType != "" {
			conditions[i] = fmt.Sprintf(" AND (ad.adv_type=%q AND ad.adv_start<=%q AND ad.adv_end>=%q and  ad.adv_status=1", advType, currentDate, currentDate)
			i = i + 1
		} else {
			conditions[i] = fmt.Sprintf(" AND (ad.adv_start<=%q AND ad.adv_end>=%q and  ad.adv_status=1", currentDate, currentDate)
			i = i + 1
		}

	}

	switch zone.ZoneType {
	case "banner":
		conditions[i] = fmt.Sprintf(" AND ad.creative_unit_type='banner' AND ad.adv_width=%q AND ad.adv_height=%q)", zone.ZoneWidth, zone.ZoneHeight)
		i = i + 1
		break
	case "interstitial":
		conditions[i] = fmt.Sprintf(" AND ad.creative_unit_type='interstitial'")
		i = i + 1
		//尺寸匹配
		if screenCode != "" {
			screenWidth, screenHeight := getScreenSize(screenCode)
			conditions[i] = fmt.Sprintf(" AND ad.adv_width=:adv_width: AND ad.adv_height=:adv_height:)", screenWidth, screenHeight)
		} else {
			conditions[i] = ")"
		}
		i = i + 1
		break
	case "mini_interstitial":

		conditions[i] = fmt.Sprintf(" AND ad.creative_unit_type='mini_interstitial' AND ad.adv_width=%q AND ad.adv_height=%q)", zone.ZoneWidth, zone.ZoneHeight)
		i = i + 1

		break
	case "open":
		conditions[i] = fmt.Sprintf(" AND (ad.adv_start<=%q AND ad.adv_end>=%q AND ad.adv_status=1 AND ad.creative_unit_type='open'", currentDate, currentDate)
		switch advType {
		case "1":
			conditions[i] = " AND ad.adv_type = 1"
			break
		case "3":
			conditions[i] = " AND ad.adv_type = 3"
			break
		case "2": //视频
			conditions[i] = " AND (ad.adv_type = 2 OR ad.adv_type = 5)"
			break
		case "4": //zip包
			conditions[i] = " AND (ad.adv_type = 4 OR ad.adv_type = 5)"
			break
		case "5": //视频及zip包
			conditions[i] = " AND ad.adv_type = 5"
			break
		default: //默认
			conditions[i] = " AND (ad.adv_type =2 OR ad.adv_type = 4 OR ad.adv_type = 5)"
			break

		}
		i = i + 1
		break
	default:
	}

	//剩余点击量检查
	conditions[i] = " AND (c_limit.total_amount_left>=1)"
	i = i + 1

	//时段定向 todo
	return strings.Join(conditions, ""), leftGeo, leftQuality
}

func isExistTargeting(targetType, targetCode string) bool {

	return true

}

func isExistQualityTargeting(targetCodes []int) bool {

	if targetCodes != nil {

		return true
	} else {

		return false
	}

}

func getScreenSize(screenCode string) (screenWidth, screenHeight string) {
	o := orm.NewOrm()

	var rstLov *Lov
	err := o.Raw("SELECT value FROM md_lov WHERE key = ? and code = ? limit 1", "screen_type", screenCode).QueryRow(&rstLov)

	if err == nil && rstLov != nil {
		arySize := strings.Split(rstLov.Value, "x")
		screenWidth = arySize[0]
		screenHeight = arySize[1]
	}

	return screenWidth, screenHeight

}

func composeString(ids []int) (retString string) {
	length := len(ids)
	for i, v := range ids {
		if i < length-1 {
			retString += fmt.Sprintf("%d,", v)
		} else {
			retString += fmt.Sprintf("%d", v)
		}

	}

	return retString
}

func LaunchQuery(provinceCode string, cityCode string, zone *Zone, qualityIds []int, advType string, screenCode string) (adUnit *AdUnits) {
	strQuery, leftGeo, leftQuality := buildQuery(provinceCode, cityCode, zone, qualityIds, advType, screenCode)

	campaignTableName := "md_campaigns"

	sql := fmt.Sprintf("SELECT Campaigns.campaign_id AS campaign_id,Campaigns.creative_show_rule AS creative_show_rule,Campaigns.campaign_priority AS campaign_priority,Campaigns.campaign_type AS campaign_type FROM %s AS Campaigns", campaignTableName)
	if leftGeo {
		sql += " LEFT JOIN md_campaign_targeting AS c1 ON Campaigns.campaign_id=c1.campaign_id"
	}

	sql += " LEFT JOIN md_campaign_targeting AS c3 ON Campaigns.campaign_id=c3.campaign_id"
	if leftQuality {
		sql += " LEFT JOIN md_campaign_targeting AS c7 ON Campaigns.campaign_id=c7.campaign_id"
	}

	sql += " LEFT JOIN md_campaign_limit AS c_limit ON Campaigns.campaign_id = c_limit.campaign_id LEFT JOIN md_ad_units AS ad ON Campaigns.campaign_id = ad.campaign_id WHERE " + strQuery

	beego.Debug(sql)

	o := orm.NewOrm()
	var campaigns []Campaigns
	num, err := o.Raw(sql).QueryRows(&campaigns)
	if err != nil || num == 0 {
		beego.Error(err)
		return nil
	}

	return nil
}
