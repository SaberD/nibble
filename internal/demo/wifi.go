package demo

// WiFiHosts defines a dense demo dataset for wlan0 to stress long result lists.
var WiFiHosts = []Host{
	{IP: "10.0.0.2", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.3", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.4", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.5", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.6", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.7", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.8", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.9", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.10", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.11", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.12", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.13", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.14", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.15", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.16", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.17", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.18", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.19", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.20", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.21", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.22", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.23", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.24", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.25", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.26", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.27", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.28", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.29", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.30", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.31", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.32", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.33", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.34", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.35", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.36", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.37", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.38", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.39", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.40", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.41", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.42", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.43", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.44", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.45", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	{IP: "10.0.0.46", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},

	{IP: "10.0.0.47", Hardware: "3c:52:82:10:20:30", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.2"}, {80, "nginx"}, {443, ""}, {3000, "Grafana"}, {9090, "Prometheus"}}},
	{IP: "10.0.0.48", Hardware: "dc:a6:32:11:22:31", Ports: []Port{{22, "SSH-2.0-dropbear"}, {80, "GoAhead-Webs"}, {443, ""}, {554, "RTSP/1.0"}, {8000, "DVR HTTP"}}},
	{IP: "10.0.0.49", Hardware: "f4:f2:6d:12:23:32", Ports: []Port{{22, "SSH-2.0-OpenSSH_8.8"}, {80, "Apache/2.4.57"}, {443, ""}, {8080, "Tomcat/10.1"}}},
	{IP: "10.0.0.50", Hardware: "70:ee:50:13:24:33", Ports: []Port{{53, "dnsmasq 2.89"}, {67, ""}, {80, "OpenWrt uhttpd"}, {443, ""}, {1900, ""}}},
	// {IP: "10.0.0.51", Hardware: "b8:27:eb:14:25:34", Ports: []Port{{22, "SSH-2.0-OpenSSH_9.0"}, {80, "Home Assistant"}, {443, ""}, {1883, "MQTT"}, {8123, "Home Assistant API"}}},
}
