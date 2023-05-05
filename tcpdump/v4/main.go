package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
)

func main() {
	// Open the TCPdump file
	handle, err := pcap.OpenOffline("C:\\Users\\ihse1\\Downloads\\http9001all-temp.pcap")
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through the packets in the file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Extract the Ethernet layer
		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethLayer == nil {
			continue
		}
		//eth, _ := ethLayer.(*layers.Ethernet)

		// Extract the IP layer
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			continue
		}
		//ip, _ := ipLayer.(*layers.IPv4)

		// Extract the TCP layer
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			continue
		}
		tcp, _ := tcpLayer.(*layers.TCP)

		// Print the packet timestamp and length
		//fmt.Printf("Packet timestamp: %v\n", packet.Metadata().Timestamp)
		//fmt.Printf("Packet length: %d\n", packet.Metadata().Length)
		//
		//// Print the Ethernet, IP, and TCP headers
		//fmt.Printf("Ethernet header: %v\n", eth)
		//fmt.Printf("IP header: %v\n", ip)
		//fmt.Printf("TCP header: %v\n", tcp)
		fmt.Println(tcp.ACK)
		// Print the packet payload
		payload := packet.ApplicationLayer()
		if payload != nil {
			fmt.Printf("Payload: %s\n", string(payload.Payload()))
			fmt.Println(payload.LayerType())
		}
	}
}
