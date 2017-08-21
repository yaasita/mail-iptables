package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
)

const ipjson = "/var/www/example/mail-iptables/tmp/ip.json"
const outfile = "/etc/network/accept_ip.sh"

type IpTable struct {
	Ip   string
	Time string
}
type IpTables []IpTable

func (ipt IpTables) Len() int {
	return len(ipt)
}
func (ipt IpTables) Less(i, j int) bool {
	return ipt[i].Time < ipt[j].Time
}
func (ipt IpTables) Swap(i, j int) {
	ipt[i], ipt[j] = ipt[j], ipt[i]
}

func main() {
	var ips IpTables
	{
		ip := make(map[string]string)
		data, err := ioutil.ReadFile(ipjson)
		check(err, "json read error")
		err = json.Unmarshal(data, &ip)
		check(err, "json decode error")
		for i, v := range ip {
			it := IpTable{
				Ip:   i,
				Time: v,
			}
			ips = append(ips, it)
		}
		sort.Sort(ips)
	}
	fp, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	check(err, "out file open error")
	fp.WriteString(`#!/bin/bash

iptables -L -n | grep -q ACCEPT_IP
if [ $? = 0 ]
then
    echo "clear ACCEPT_IP chain"
    iptables -F ACCEPT_IP
else
    echo "define ACCEPT_IP chain"
    iptables -N ACCEPT_IP
fi

# iij
`)
	for _, v := range ips {
		fp.WriteString(fmt.Sprintf("iptables -A ACCEPT_IP -s %15s -j RETURN # %s\n", v.Ip, v.Time))
	}
	fp.WriteString(`
# DROP
iptables -A ACCEPT_IP -p tcp --dport 993 -j LOG --log-prefix iptables --log-level=info
iptables -A ACCEPT_IP -j DROP
`)
	fp.Close()
	err = exec.Command(outfile).Run()
	check(err, "command error")
}

func check(e error, s string) {
	if e != nil {
		fmt.Fprintln(os.Stderr, s)
		panic(e)
	}
}
