package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ScanResult struct {
	IP    net.IP
	Open  bool
	Error error
}

func main() {
	// Define the IP range to scan
	ipRange := "192.168.0.1-10"

	// Split the IP range into start and end IP addresses
	startIP, endIP, err := parseIPRange(ipRange)
	if err != nil {
		fmt.Println("Invalid IP range:", err)
		return
	}

	// Perform the IP scan
	fmt.Println("Scanning IP range:", ipRange)
	results := scanIPRange(startIP, endIP)

	// Print the scan results
	fmt.Println("\nScan Results:")
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("Error scanning IP %s: %s\n", result.IP, result.Error)
			continue
		}

		status := "Closed"
		if result.Open {
			status = "Open"
		}
		fmt.Printf("%s: %s\n", result.IP, status)
	}
}

func parseIPRange(ipRange string) (net.IP, net.IP, error) {
	ips := net.ParseIP(ipRange)
	if ips != nil {
		return ips, ips, nil
	}

	ip, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, nil, err
	}

	startIP := ip.Mask(ipNet.Mask)
	endIP := make(net.IP, len(startIP))
	copy(endIP, startIP)
	for i := len(endIP) - 1; i < len(endIP); i++ {
		endIP[i] |= ^ipNet.Mask[i]
	}

	return startIP, endIP, nil
}

func scanIPRange(startIP, endIP net.IP) []*ScanResult {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var results []*ScanResult

	for ip := startIP; compareIP(ip, endIP) <= 0; incrementIP(ip) {
		wg.Add(1)
		go func(ip net.IP) {
			defer wg.Done()
			result := ScanResult{IP: ip}
			if isPortOpen(ip, 80, time.Second) {
				result.Open = true
			}
			mutex.Lock()
			results = append(results, &result)
			mutex.Unlock()
		}(ip.To4())
	}

	wg.Wait()
	return results
}

func compareIP(ip1, ip2 net.IP) int {
	ip1 = ip1.To4()
	ip2 = ip2.To4()
	return bytesToUint(ip1) - bytesToUint(ip2)
}

func bytesToUint(ip net.IP) int {
	return int(ip[0])<<24 | int(ip[1])<<16 | int(ip[2])<<8 | int(ip[3])
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

func isPortOpen(ip net.IP, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip.String(), port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
