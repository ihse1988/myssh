package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	// Open the TCPdump file
	handle, err := pcap.OpenOffline("tcpdump.pcap")
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through the packets in the file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	var req *http.Request
	var resp *http.Response
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

		// Check if this is an HTTP request
		if tcp.SrcPort == layers.TCPPort(80) {
			// Parse the HTTP request
			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(tcp.Payload)))
			if err != nil {
				log.Println(err)
				continue
			}
		}

		// Check if this is an HTTP response
		if tcp.DstPort == layers.TCPPort(80) {
			// Parse the HTTP response
			resp, err = http.ReadResponse(bufio.NewReader(bytes.NewReader(tcp.Payload)), req)
			if err != nil {
				log.Println(err)
				continue
			}

			// Combine the HTTP request and response
			httpTransaction := &bytes.Buffer{}
			fmt.Fprintf(httpTransaction, "%s %s %s\n", req.Method, req.URL.String(), req.Proto)
			req.Header.Write(httpTransaction)
			httpTransaction.WriteString("\r\n")
			resp.Write(httpTransaction)

			// Write the combined transaction to a file
			file, err := os.Create("http_transaction.txt")
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			file.Write(httpTransaction.Bytes())
		}
	}
}
