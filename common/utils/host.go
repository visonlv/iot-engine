package utils

import (
	"net"
	"strings"

	"github.com/visonlv/go-vkit/logger"
)

func GetLocalIp() (string, error) {
	// 先从网卡获取
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorf("GetLocalIp fail to get net interfaces ipAddress: %v\n", err)
		return "", err
	}

	ips := make([]string, 0)
	for _, address := range interfaceAddr {
		ipNet, isVailIpNet := address.(*net.IPNet)
		// 检查ip地址判断是否回环地址
		if isVailIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	if len(ips) == 1 {
		logger.Infof("GetLocalIp get ip from interface:%v", ips)
		return ips[0], nil
	} else {
		logger.Infof("GetLocalIp get ip from interface fail, more then one ip:%v", ips)
	}

	//域名发现获取
	dnsServer := "8.8.8.8:80"
	logger.Infof("GetLocalIp get ip from dns:%v", dnsServer)
	conn, err := net.Dial("udp", dnsServer)
	if err != nil {
		logger.Infof("GetLocalIp get ip from dns:%s fail:%s", dnsServer, err)
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	ip := localAddr[0:idx]
	logger.Infof("GetLocalIp get ip from dns:%v success ip:%s", dnsServer, ip)
	return ip, nil
}
