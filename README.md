# hitron-exporter

Added Inital support for Hitron CGN3

Metrics exporter for Hitron CGNV4-FX2 routers, which are (used to be?) distributed as part of Unitymedia/Vodafone Business cable internet.


## Running

```bash
docker run -it --rm -p 9101:80 ghcr.io/cfstras/hitron-exporter:latest --host --pass XYZ
```

### docker-compose

```yaml
hitron_exporter:
  image: ghcr.io/cfstras/hitron-exporter:latest
  command:
    - --pass=mySecretPassword
#   - --user=admin
#   - --host=http://192.168.0.1:80/
  ports:
    - 9101:80
  restart: unless-stopped
```

### Example output

```bash
➜ curl localhost/metrics
# HELP hitron_address Hardware and IP Addresses in labels
# TYPE hitron_address gauge
hitron_address{lan_ip="192.168.0.1/24",rf_mac="68:XX:YY:ZZ:AA:BB",wan_ip="1.2.3.4/21"} 1
# HELP hitron_cm_bpi_status DOCSIS Provisioning BPI Status
# TYPE hitron_cm_bpi_status gauge
hitron_cm_bpi_status{auth="authorized",tek="operational"} 1
# HELP hitron_cm_dhcp_lease_duration DOCSIS DHCP Lease duration
# TYPE hitron_cm_dhcp_lease_duration counter
hitron_cm_dhcp_lease_duration 259200
# HELP hitron_cm_dhcp_success DOCSIS Provisioning DHCP Status
# TYPE hitron_cm_dhcp_success gauge
hitron_cm_dhcp_success 1
# HELP hitron_cm_docsis_addr DOCSIS IP Addresses
# TYPE hitron_cm_docsis_addr gauge
hitron_cm_docsis_addr{gateway="10.40.XX.YY",ip="10.40.XX.YY",netmask="255.255.240.0"} 1
# HELP hitron_cm_download_config_success DOCSIS Provisioning Download CM Config File Status
# TYPE hitron_cm_download_config_success gauge
hitron_cm_download_config_success 1
# HELP hitron_cm_find_downstream_success DOCSIS Provisioning Lock Downstream Status
# TYPE hitron_cm_find_downstream_success gauge
hitron_cm_find_downstream_success 1
# HELP hitron_cm_hwinit_success DOCSIS Provisioning HWInit Status
# TYPE hitron_cm_hwinit_success gauge
hitron_cm_hwinit_success 1
# HELP hitron_cm_network_access_status DOCSIS Network Access Permission
# TYPE hitron_cm_network_access_status gauge
hitron_cm_network_access_status 1
# HELP hitron_cm_ranging_success DOCSIS Provisioning Ranging Status
# TYPE hitron_cm_ranging_success gauge
hitron_cm_ranging_success 1
# HELP hitron_cm_registration_success DOCSIS Provisioning Registration Status
# TYPE hitron_cm_registration_success gauge
hitron_cm_registration_success 1
# HELP hitron_info_uptime System uptime
# TYPE hitron_info_uptime counter
hitron_info_uptime 525245
# HELP hitron_lan_device LAN Device table
# TYPE hitron_lan_device gauge
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.12",ip_type="static",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 1
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.2",ip_type="static",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 1
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.20",ip_type="dhcp",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 1
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.21",ip_type="dhcp",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 1
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.22",ip_type="dhcp",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 1
hitron_lan_device{comnum="1",interface="Ethernet",ip="192.168.0.26",ip_type="dhcp",ip_version="IPv4",mac="XX:YY:ZZ:AA:BB:CC"} 0
# HELP hitron_login_success_bool 1 if the login was successful
# TYPE hitron_login_success_bool gauge
hitron_login_success_bool 1
# HELP hitron_scrape_time Time the scrape run took
# TYPE hitron_scrape_time gauge
hitron_scrape_time{component="CMDocsisWAN"} 0.281379708
hitron_scrape_time{component="CMInit"} 0.372671959
hitron_scrape_time{component="ConnectInfo"} 1.81596975
hitron_scrape_time{component="DownstreamInfo"} 0.24948125
hitron_scrape_time{component="Info"} 0.610619958
hitron_scrape_time{component="UpstreamInfo"} 0.127571375
hitron_scrape_time{component="all"} 4.880440167
hitron_scrape_time{component="login"} 0.898155542
# HELP hitron_traffic Basic traffic counters. if=wan/lan, dir=send/recv.
# TYPE hitron_traffic counter
hitron_traffic{dir="recv",if="lan"} 1.04112062464e+09
hitron_traffic{dir="recv",if="wan"} 9.5321849856e+08
hitron_traffic{dir="send",if="lan"} 1.83609851904e+09
hitron_traffic{dir="send",if="wan"} 6.4627933184e+08
# HELP hitron_version Versions in labels
# TYPE hitron_version gauge
hitron_version{hw_version="1A",serial="VCA123456",sw_version="4.1.2.3-SNIP"} 1
```
## License

See [LICENSE.md](LICENSE.md)  
Heavily inspired by and structure stolen from [fluepke's vodafone-station-exporter](https://github.com/Fluepke/vodafone-station-exporter).  
