package demo

type Port struct {
	Port   int
	Banner string
}

type Host struct {
	IP       string
	Hardware string
	Ports    []Port
}

// Hosts defines fake hosts with real MAC addresses so demo uses the OUI lookup.
var Hosts = []Host{
	{
		IP: "192.168.1.1", Hardware: "f0:9f:c2:1a:22:01",
		Ports: []Port{
			{22, "SSH-2.0-OpenSSH_8.4"},
			{80, "UniFi OS 3.2.12"},
			{443, ""},
		},
	},
	{
		IP: "192.168.1.50", Hardware: "48:b0:2d:5e:a3:10",
		Ports: []Port{
			{22, "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6"},
			{5432, ""},
		},
	},
	{
		IP: "192.168.1.75", Hardware: "9c:b7:0d:0a:3f:12",
		Ports: []Port{
			{22, "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.11"},
		},
	},
	{
		IP: "192.168.1.100", Hardware: "f0:ee:7a:ab:cd:ef",
		Ports: []Port{
			{80, "Apache/2.4.56"},
			{443, ""},
			{8080, "Jetty 11.0.15"},
		},
	},
	{
		IP: "10.0.0.42", Hardware: "d8:3a:dd:11:22:33",
		Ports: []Port{
			{22, "SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u2"},
			{80, "lighttpd/1.4.69"},
			{1883, ""},
		},
	},
}
