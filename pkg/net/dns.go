package net

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bogdanovich/dns_resolver"
	"strings"
)

//GetMasterAddresses Get preferred outbound ip of this machine
func GetMasterAddresses(dcName string, clusterName string, domainName string, size int, svcIP string) string {

	resolver, err := dns_resolver.NewFromResolvConf("/etc/resolv.conf")

	if err != nil {
		log.Fatal(err)
	}
	// In case of i/o timeout
	resolver.RetryTimes = 5

	// loop through all the masters
	// lookup the IPs
	// append to a slice which is returned as a CSV
	var ips []string
	i := 1
	for i <= size {
		hostname := fmt.Sprintf("%s-%smaster-%d.%s", dcName, clusterName, i, domainName)
		log.Debug("Looking up host: ", hostname)
		addresses, err := resolver.LookupHost(hostname)
		if err != nil {
			log.Fatal("Error resolving hostname: ", err)
		}
		ips = append(ips, addresses[0].String())
		i++
	}

	ips = append(ips, svcIP)

	return strings.Join(ips, ",")

}
