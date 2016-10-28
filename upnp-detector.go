package main

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"
)

const (
	srvAddr         = "239.255.255.250:1900"
	maxDatagramSize = 8192
)

var (
	confDirs = [3]string{"/etc/upnp-detector/", "$HOME/.upnp-detector", "."}
	conn     *net.UDPConn
	dev      string
	host     string
	port     string
)

func main() {

	gaddr, err := net.ResolveUDPAddr("udp4", srvAddr)
	checkError(err)
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1900")

	viper.SetConfigName("upnp-detector") // will match upnp-detector.{toml,json} etc.
	for _, dir := range confDirs {
		viper.AddConfigPath(dir)
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	port = viper.GetString("host")
	port = viper.GetString("port")
	dev = viper.GetString("device")
	//iface, err := net.InterfaceByName(dev)
	checkError(err)

	conn, err = net.ListenMulticastUDP("udp4", nil, gaddr)
	checkError(err)
	conn.SetReadBuffer(maxDatagramSize)

	for {
		handleConnection(conn)
	}
}

func handleConnection(c *net.UDPConn) {
	buf := make([]byte, maxDatagramSize)
	_, addr, err := c.ReadFromUDP(buf)
	checkError(err)

	fmt.Println("Received upnp request from ", addr)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
