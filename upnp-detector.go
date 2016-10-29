package main

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	srvAddr         = "239.255.255.250:1900"
	maxDatagramSize = 8192
	endpoint        = "/upnp/record"
)

var (
	confDirs = [3]string{"/etc/upnp-detector/", "$HOME/.upnp-detector", "."}
	conn     *net.UDPConn
	dev      string
	host     string
	port     string
	transp   *http.Transport
	client   *http.Client
	url      string
	username string
	password string
)

type upnpRequest struct {
	SrcIP   string
	Headers []string
}

func main() {

	gaddr, err := net.ResolveUDPAddr("udp4", srvAddr)
	checkError(err)
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "9090")
	viper.SetDefault("device", nil)

	viper.SetConfigName("upnp-detector") // will match upnp-detector.{toml,json} etc.
	for _, dir := range confDirs {
		viper.AddConfigPath(dir)
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}
	host = viper.GetString("host")
	port = viper.GetString("port")
	dev = viper.GetString("device")
	username = viper.GetString("username")
	password = viper.GetString("password")

	iface, err := net.InterfaceByName(dev)
	checkError(err)

	conn, err = net.ListenMulticastUDP("udp4", iface, gaddr)
	checkError(err)
	conn.SetReadBuffer(maxDatagramSize)
	// TODO add config flags for TLS
	// TODO add config for http auth client config
	//transp = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	url = "http://" + host + ":" + port + endpoint
	client = &http.Client{}

	for {
		// read one UPnP packet
		buf := make([]byte, maxDatagramSize)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Could not read from UPD connection: " + err.Error())
			continue
		}
		go handlePacket(buf[:n], addr)
	}
}

func handlePacket(buf []byte, addr *net.UDPAddr) {

	// serialize and discard the last two headers as they will be empty
	headers := strings.Split(string(buf), "\r\n")
	upnp := upnpRequest{
		SrcIP:   addr.IP.String(),
		Headers: headers[:len(headers)-2]}

	j, err := json.Marshal(upnp)
	if err != nil {
		fmt.Println("Could not marshal to json: " + err.Error())
		return
	}

	// send the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	if err != nil {
		fmt.Println("Could not create request : " + err.Error())
		return
	}
	//req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not Post request: " + err.Error())
		return
	}
	defer resp.Body.Close()

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
