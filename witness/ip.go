package witness

import (
	"net"
	"fmt"
	logger "log"
)


func findIpWithStdin() {
	getLocalIp(func (ip *net.IP) bool {
		var cf int
		fmt.Print("Local IP is: ", ip, " ? (y/N) ")
		fmt.Scanf("%c\n", &cf)
		if cf == 'y' {
			c.Host = ip.String()
			saveConfig(c, *configFile)
			return true;
		}
		return false
	})
}


/**
 * 如果 setter 返回 true, 则终止 ip 地址的便利
 */
func getLocalIp(setter func(*net.IP) bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Fatalln("Cannot get Network Interfaces", err)
	}
	
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.Fatalln("Cannot get Network Address", err)
		}
		
		for _, addr := range addrs {
			var ip net.IP
			// log("addr=", addr)
			switch v := addr.(type) {
				case *net.IPNet:
								ip = v.IP
				case *net.IPAddr:
								ip = v.IP
			}
			if ip != nil 				&&
				 ip.To4() != nil 	&&
				 !ip.IsLoopback() &&
				 !ip.IsUnspecified() {
				if setter(&ip) {
					return
				}
			}
		}
	}
}


func findIpWithConfig() bool {
	var isfind bool

	getLocalIp(func (ip *net.IP) bool {
		if c.Host == ip.String() {
			isfind = true
			return true
		}
		return false
	})
	return isfind
}