package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	// ... other imports
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	itemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

func performScan(subnet string) []string {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var wg sync.WaitGroup
	hosts := make([]string, 0)
	var mutex sync.Mutex

	incrementIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", ip+":80", 2*time.Second)
			if err == nil {
				conn.Close()
				hostInfo := ip

				names, err := net.LookupAddr(ip)
				if err == nil && len(names) > 0 {
					hostInfo += ", " + names[0]
				}

				mutex.Lock()
				hosts = append(hosts, hostInfo)
				mutex.Unlock()
			}
		}(ip.String())
	}

	wg.Wait()
	return hosts
}

func prettyPrintResults(results []string) {
	title := titleStyle.Render("Scan Results:")
	fmt.Println(title)

	for _, result := range results {
		item := itemStyle.Render(result)
		fmt.Println(" •", item)
	}
}
