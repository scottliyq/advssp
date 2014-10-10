package models

type AdUnits struct {
	AdvId                             int
	CampaignId                        int
	AdvType                           string
	AdvStatus                         string
	AdvClickUrlType                   string
	AdvClickUrl                       string
	AdvClickOpenType                  string
	AdvChtml                          string
	AdvMraid                          string
	AdvImpressionTrackingUrl          string
	AdvImpressionTrackingUrlIresearch string
	AdvImpressionTrackingUrlAdmaster  string
	AdvImpressionTrackingUrlNielsen   string
	AdvCreativeUrl                    string
	AdvCreativeUrl2                   string
	AdvCreativeUrl3                   string
}

type Zone struct {
	EntryId         int
	PublicationId   int
	PlacementHash   string
	ZoneLastrequest int
	ZoneType        string
	ZoneWidth       string
	ZoneHeight      string
}

type Lov struct {
	ScreenType string
	Code       string
	Value      string
}

type Device struct {
	DeviceId      int
	DeviceMovment string
	DeviceName    string
}

type DevicePackageMatrix struct {
	DeviceId  int
	PackageId int
}

type RegionalTargeting struct {
	TargetingCode string
	RegionName    string
}

type Campaigns struct {
	CreativeShowRule string
	CampaignId       int
	Priority         int
	Type             string
}
