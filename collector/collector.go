package collector

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Collector struct {
	Router *HitronRouter
}

const prefix = "hitron_"

var (
	loginSuccessDesc *prom.Desc = prom.NewDesc(
		prefix+"login_success_bool", "1 if the login was successful", nil, nil)
	scrapeTimeDesc *prom.Desc = prom.NewDesc(
		prefix+"scrape_time", "Time the scrape run took", []string{"component"}, nil)

	// SysInfo
	systemLanUptimeDesc = prom.NewDesc(
		prefix+"info_lan_uptime", "System uptime", nil, nil)
	systemWanUptimeDesc = prom.NewDesc(
		prefix+"info_wan_uptime", "System uptime", nil, nil)
	versionDesc = prom.NewDesc(
		prefix+"version", "Versions in labels",
		[]string{"hw_version", "sw_version", "serial"}, nil)
	addressDesc = prom.NewDesc(
		prefix+"address", "Hardware and IP Addresses in labels",
		[]string{"wan_ip", "lan_ip", "rf_mac"}, nil)
	trafficDesc = prom.NewDesc(
		prefix+"traffic", "Basic traffic counters. if=wan/lan, dir=send/recv.",
		[]string{"if", "dir"}, nil)

	// CMInit
	cmHwInitDesc = prom.NewDesc(
		prefix+"cm_hwinit_success", "DOCSIS Provisioning HWInit Status", nil, nil)
	cmFindDownstreamDesc = prom.NewDesc(
		prefix+"cm_find_downstream_success", "DOCSIS Provisioning Lock Downstream Status", nil, nil)
	cmRangingDesc = prom.NewDesc(
		prefix+"cm_ranging_success", "DOCSIS Provisioning Ranging Status", nil, nil)
	cmDhcpDesc = prom.NewDesc(
		prefix+"cm_dhcp_success", "DOCSIS Provisioning DHCP Status", nil, nil)
	cmDownloadConfigDesc = prom.NewDesc(
		prefix+"cm_download_config_success", "DOCSIS Provisioning Download CM Config File Status", nil, nil)
	cmRegistrationDesc = prom.NewDesc(
		prefix+"cm_registration_success", "DOCSIS Provisioning Registration Status", nil, nil)
	cmBPIDesc = prom.NewDesc(
		prefix+"cm_bpi_status", "DOCSIS Provisioning BPI Status",
		[]string{"auth", "tek"}, nil)
	cmNetworkAccessDesc = prom.NewDesc(
		prefix+"cm_network_access_status", "DOCSIS Network Access Permission", nil, nil)

	// CMDocsisWAN
	cmDocsisAddressDesc = prom.NewDesc(
		prefix+"cm_docsis_addr", "DOCSIS IP Addresses",
		[]string{"ip", "netmask", "gateway"}, nil)
	cmIpLeaseDurationDesc = prom.NewDesc(
		prefix+"cm_dhcp_lease_duration", "DOCSIS DHCP Lease duration", nil, nil)

	// ConnectInfo
	lanDeviceDesc = prom.NewDesc(
		prefix+"lan_device", "LAN Device table",
		[]string{"ip", "ip_version", "mac", "ip_type", "interface", "comnum"}, nil)
)

func (c *Collector) Describe(ch chan<- *prom.Desc) {
	ch <- loginSuccessDesc
	ch <- scrapeTimeDesc

	// SysInfo
	ch <- systemLanUptimeDesc
	ch <- systemWanUptimeDesc
	ch <- versionDesc
	ch <- addressDesc
	ch <- trafficDesc

	// CMInit
	ch <- cmHwInitDesc
	ch <- cmFindDownstreamDesc
	ch <- cmRangingDesc
	ch <- cmDhcpDesc
	ch <- cmDownloadConfigDesc
	ch <- cmRegistrationDesc
	ch <- cmBPIDesc
	ch <- cmNetworkAccessDesc

	// CMDocsisWAN
	ch <- cmDocsisAddressDesc
	ch <- cmIpLeaseDurationDesc

	// ConnectInfo
	ch <- lanDeviceDesc

}
func (c *Collector) Collect(ch chan<- prom.Metric) {
	defer measureTime(ch, "all")()

	loginFinished := measureTime(ch, "login")
	session, err := c.Router.Login()
	if err != nil {
		ch <- prom.MustNewConstMetric(loginSuccessDesc, prom.GaugeValue, 0)
		return
	}
	defer session.Logout()
	ch <- prom.MustNewConstMetric(loginSuccessDesc, prom.GaugeValue, 1)
	loginFinished()

	var wg sync.WaitGroup
	wg.Add(6)

	// these could be run in parallel, but the webserver seems to be serial
	// so that only screws up our section timing metrics
	c.CollectInfo(&wg, session, ch)
	c.CollectCMInit(&wg, session, ch)
	c.CollectCMDocisWAN(&wg, session, ch)
	c.CollectConnectInfo(&wg, session, ch)
	c.CollectDonwstreamInfo(&wg, session, ch)
	c.CollectUpstreamInfo(&wg, session, ch)

	wg.Wait()
	log.Debug("Collect() done.")
}

func (c *Collector) CollectInfo(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "Info")()
	defer wg.Done()

	info, err := session.Info()
	if err != nil {
		// todo count errors
		log.Info("Info: ", err)
		return
	}
	ch <- prom.MustNewConstMetric(systemLanUptimeDesc, prom.CounterValue, parseDuration(info.SystemLANUptime))
	ch <- prom.MustNewConstMetric(systemWanUptimeDesc, prom.CounterValue, parseDuration(info.SystemWANUptime))
	ch <- prom.MustNewConstMetric(versionDesc, prom.GaugeValue, 1, info.HwVersion, info.SwVersion, info.SerialNumber)
	ch <- prom.MustNewConstMetric(addressDesc, prom.GaugeValue, 1, info.WanIp, info.LanIp, info.RfMac)
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.LRecPkt), "lan", "recv")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.LSendPkt), "lan", "send")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.WRecPkt), "wan", "recv")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.WSendPkt), "wan", "send")
}

func (c *Collector) CollectCMInit(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "CMInit")()
	defer wg.Done()

	cmInit, err := session.CMInit()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}
	ch <- prom.MustNewConstMetric(cmHwInitDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.HwInit))
	ch <- prom.MustNewConstMetric(cmFindDownstreamDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.FindDownstream))
	ch <- prom.MustNewConstMetric(cmRangingDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.Ranging))
	ch <- prom.MustNewConstMetric(cmDhcpDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.Dhcp))
	ch <- prom.MustNewConstMetric(cmDownloadConfigDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.DownloadCfg))
	ch <- prom.MustNewConstMetric(cmRegistrationDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.Registration))
	ch <- prom.MustNewConstMetric(cmNetworkAccessDesc, prom.GaugeValue,
		is(NetworkAccessPermitted, cmInit.NetworkAccess))

	bpiDesc := map[string]string{}
	kvs := strings.Split(cmInit.BpiStatus, ", ")
	for _, kv := range kvs {
		split := strings.SplitN(kv, ":", 2)
		if len(split) == 2 {
			bpiDesc[split[0]] = split[1]
		}
	}
	ch <- prom.MustNewConstMetric(cmBPIDesc, prom.GaugeValue,
		1, bpiDesc["AUTH"], bpiDesc["TEK"])
}

func (c *Collector) CollectCMDocisWAN(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "CMDocsisWAN")()
	defer wg.Done()

	wan, err := session.CMDocsisWAN()
	if err != nil {
		log.Info("CMDocsisWAN: ", err)
		return
	}
	ch <- prom.MustNewConstMetric(cmDocsisAddressDesc, prom.GaugeValue,
		is(NetworkAccessPermitted, wan.NetworkAccess),
		wan.CmIpAddress, wan.CmNetMask, wan.CmGateway)
	ch <- prom.MustNewConstMetric(cmIpLeaseDurationDesc, prom.CounterValue,
		parseLeaseDuration(wan.CmIpLeaseDuration))
}

func (c *Collector) CollectConnectInfo(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "ConnectInfo")()
	defer wg.Done()

	info, err := session.ConnectInfo()
	if err != nil {
		log.Info("ConnectInfo: ", err)
		return
	}
	for _, device := range info {
		connectType := string(device.ConnectType)
		if device.ConnectType == DHCP {
			connectType = "dhcp"
		} else if device.ConnectType == Static {
			connectType = "static"
		}
		ch <- prom.MustNewConstMetric(lanDeviceDesc, prom.GaugeValue, is("active", device.Online),
			device.IpAddr, device.IpType, device.MacAddr, connectType, device.Interface, fmt.Sprint(device.Comnum))
	}
}

func (c *Collector) CollectUpstreamInfo(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "UpstreamInfo")()
	defer wg.Done()

	_, err := session.UpstreamInfo()
	if err != nil {
		log.Info("UpstreamInfo: ", err)
		return
	}
}

func (c *Collector) CollectDonwstreamInfo(wg *sync.WaitGroup, session *Session, ch chan<- prom.Metric) {
	defer measureTime(ch, "DownstreamInfo")()
	defer wg.Done()

	_, err := session.DownstreamInfo()
	if err != nil {
		log.Info("DownstreamInfo: ", err)
		return
	}
}

func is(expected, actual string) float64 {
	if expected == actual {
		return 1
	}
	return 0
}

// parseDuration parses a duration in the format of "05 Days,21 Hours,33 Minutes,44 Seconds"
// Example
func parseLeaseDuration(raw string) float64 {
	var days, hours, minutes, seconds time.Duration
	_, err := fmt.Sscanf(raw, "D: %2d H: %2d M: %2d S:%2d", &days, &hours, &minutes, &seconds)
	if err != nil {
		log.Warn("Unknown uptime format: ", err, " ", raw)
		return -1
	}

	return float64(seconds +
		minutes*60 +
		hours*3600 +
		days*24*3600)
}

func parseDuration(raw string) float64 {
	var days, hours, minutes, seconds time.Duration
	_, err := fmt.Sscanf(raw, "%3d days %2dh:%2dm:%2ds", &days, &hours, &minutes, &seconds)
	if err != nil {
		log.Warn("Unknown uptime format: ", err, " ", raw)
		return -1
	}

	return float64(seconds +
		minutes*60 +
		hours*3600 +
		days*24*3600)
}

var scaleMap = map[string]float64{
	"K": 1024,
	"M": math.Pow(1024, 2),
	"G": math.Pow(1024, 3),
	"T": math.Pow(1024, 4),
	"E": math.Pow(1024, 5),
}

func parsePkt(raw string) float64 {
	var out float64
	var scale string
	_, err := fmt.Sscanf(raw, "%f%1s Bytes", &out, &scale)
	if err != nil {
		return -1
	}
	factor, ok := scaleMap[scale]
	if !ok {
		log.Warn("Unknown scale factor in ", raw)
		return -1
	}
	return out * factor
}

func measureTime(ch chan<- prom.Metric, label string) func() {
	startTime := time.Now()

	return func() {
		duration := time.Since(startTime).Seconds()
		ch <- prom.MustNewConstMetric(scrapeTimeDesc, prom.GaugeValue, duration, label)
	}
}
