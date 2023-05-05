package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	device      string = "\\Device\\NPF_{29B2AD5D-FAC0-478A-9985-AADEBE99D52E}"
	snapshotLen int32  = 1024
	promiscuous bool   = false
	err         error
	timeout     time.Duration = 2 * time.Second
	handle      pcap.Handle
	packetCount int = 0
	// Will reuse these for each packet
	ethLayer layers.Ethernet
	ipLayer  layers.IPv4
	tcpLayer layers.TCP
)

type httpStreamFactory struct{}

func (factory *httpStreamFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	stream := &httpStream{
		netFlow: netFlow,
		tcpFlow: tcpFlow,
		r:       tcpreader.NewReaderStream(),
	}
	go stream.run()
	return stream
}

type httpStream struct {
	netFlow, tcpFlow gopacket.Flow
	r                tcpreader.ReaderStream
}

func (stream *httpStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	//TODO implement me
	panic("implement me")
}

func (stream *httpStream) ReassemblyComplete() {
	//TODO implement me
	panic("implement me")
}

func (stream *httpStream) run() {
	for {
		if err := stream.r.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			fmt.Println("Failed to set read deadline:", err)
			return
		}
		data, err := stream.r.ReadBytes('\n')
		if err != nil {
			fmt.Println("Failed to read bytes:", err)
			return
		}
		if len(data) == 0 {
			return
		}
		fmt.Printf("[%s] %s", stream.netFlow, string(data))
	}
}

func main() {
	FindAllDevs()
	//add file
	var handle *pcap.Handle
	var err error
	// Open device
	if len(os.Args) > 1 {
		device = os.Args[1]
	}
	fmt.Println(device)
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Open output pcap file and write header
	f, _ := os.Create("test.pcap")
	w := pcapgo.NewWriter(f)
	w.WriteFileHeader(uint32(snapshotLen), layers.LinkTypeEthernet)
	defer f.Close()

	// Set filter
	var filter string = "tcp and port 80"
	filter = "(tcp[((tcp[12:1] & 0xf0) >> 2):4] = 0x47455420 or tcp[((tcp[12:1] & 0xf0) >> 2):4] = 0x504f5354)"
	//filter = ""
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Only capturing TCP port 80 packets.")

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		//if packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
		//	continue
		//}
		//tcp := packet.TransportLayer().(*layers.TCP)
		//assembler.AssembleWithTimestamp(
		//	packet.NetworkLayer().NetworkFlow(),
		//	tcp,
		//	packet.Metadata().Timestamp)
		//上面代码是解决tcp重传的问题，暂时没有写完整

		// Process packet here
		//fmt.Println(packet)
		//w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		packetCount++
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if len(tcp.Payload) != 0 {
				//payloadStr, err := decodeUTF8(tcp.Payload)
				//if err != nil {
				//	// 处理解码错误
				//}
				reader := bufio.NewReader(bytes.NewReader(tcp.Payload))
				//reader := bufio.NewReader(strings.NewReader(payloadStr + "\r\n"))
				httpReq, err := http.ReadRequest(reader)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(httpReq)
			}
		}
		//printPacketInfo(packet)
		//newDecode(packet)
		// Only capture 100 and then stop
		printHTTPPacketInfo(packet)
		if packetCount > 100 {
			break
		}
	}
}

func decodeUTF8(payload []byte) (string, error) {
	// 先检查字节切片是否包含有效的 UTF-8 编码
	if !utf8.Valid(payload) {
		return "", fmt.Errorf("invalid UTF-8 encoding")
	}
	// 解码字节切片为字符串
	return string(payload), nil
}

func printHTTPPacketInfo(packet gopacket.Packet) {
	appLayer := packet.ApplicationLayer()
	if appLayer != nil {
		// Get the TCP layer
		payload := string(appLayer.Payload())
		if strings.HasPrefix(payload, "GET") || strings.HasPrefix(payload, "POST") {

			fmt.Println("--------------------------------------------------------------------")
			//fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
			//fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
			//// Ethernet type is typically IPv4 but could be ARP or other
			//fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
			//fmt.Printf("From %s to %s\n", ip.SrcIP, ip.DstIP)
			//fmt.Println("Protocol: ", ip.Protocol)
			//fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
			//fmt.Println("Sequence number: ", tcp.Seq)

			//fmt.Printf("%s\n", applicationLayer.Payload())
			//a := packet.TransportLayer().TransportFlow()
			//fmt.Println(a.EndpointType())
			//fmt.Println("HTTP found!")
			//fmt.Println(payload)
			//解释正则表达式
			reg := regexp.MustCompile(`(?s)(GET|POST) (.*?) HTTP.*Host: (.*?)\n`)
			if reg == nil {
				fmt.Println("MustCompile err")
				return
			}
			//提取关键信息
			result := reg.FindStringSubmatch(payload)
			//fmt.Println(result)
			//fmt.Println(len(result))
			//fmt.Println(result[2], result[3])
			if len(result) == 4 {
				strings.TrimSpace(result[2])
				url := "http://" + strings.TrimSpace(result[3]) + strings.TrimSpace(result[2])
				fmt.Println("url:", url)
				fmt.Println("host:", result[3])
			} else {
				fmt.Println("error===================")
			}

		}
	}
}

// new decode
func newDecode(packet gopacket.Packet) {
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&ethLayer,
		&ipLayer,
		&tcpLayer,
	)
	foundLayerTypes := []gopacket.LayerType{}

	err := parser.DecodeLayers(packet.Data(), &foundLayerTypes)
	if err != nil {
		fmt.Println("Trouble decoding layers: ", err)
	}

	for _, layerType := range foundLayerTypes {
		if layerType == layers.LayerTypeIPv4 {
			fmt.Println("IPv4: ", ipLayer.SrcIP, "->", ipLayer.DstIP)
		}
		if layerType == layers.LayerTypeTCP {
			fmt.Println("TCP Port: ", tcpLayer.SrcPort, "->", tcpLayer.DstPort)
			fmt.Println("TCP SYN:", tcpLayer.SYN, " | ACK:", tcpLayer.ACK)
			fmt.Println(string(tcpLayer.LayerPayload()))
			applicationLayer := packet.ApplicationLayer()
			if applicationLayer != nil {
				fmt.Println("Application layer/Payload found.")
				fmt.Printf("%s\n", applicationLayer.Payload())

				// Search for a string inside the payload
				if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
					fmt.Println("HTTP found!")
				}
			}
		}

	}

}

// 解析包
func printPacketInfo(packet gopacket.Packet) {
	// Let’s see if the packet is an ethernet packet
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		fmt.Println("Ethernet layer detected.")
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
		fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
		// Ethernet type is typically IPv4 but could be ARP or other
		fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
		fmt.Println()
	}

	// Let’s see if the packet is IP (even though the ether type told us)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		fmt.Println("IPv4 layer detected.")
		ip, _ := ipLayer.(*layers.IPv4)

		// IP layer variables:
		// Version (Either 4 or 6)
		// IHL (IP Header Length in 32-bit words)
		// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
		// Checksum, SrcIP, DstIP
		fmt.Printf("From %s to %s\n", ip.SrcIP, ip.DstIP)
		fmt.Println("Protocol: ", ip.Protocol)
		fmt.Println()
	}

	// Let’s see if the packet is TCP
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		fmt.Println("TCP layer detected.")
		tcp, _ := tcpLayer.(*layers.TCP)

		// TCP layer variables:
		// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
		// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
		fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
		fmt.Println("Sequence number: ", tcp.Seq)
		fmt.Println()
	}

	// Iterate over all layers, printing out each layer type
	fmt.Println("All packet layers:")
	for _, layer := range packet.Layers() {
		fmt.Println("- ", layer.LayerType())
	}

	// When iterating through packet.Layers() above,
	// if it lists Payload layer then that is the same as
	// this applicationLayer. applicationLayer contains the payload
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		fmt.Println("Application layer/Payload found.")
		fmt.Printf("%s\n", applicationLayer.Payload())

		// Search for a string inside the payload
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			fmt.Println("HTTP found!")
		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}

func FindAllDevs() {
	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// Print device information
	fmt.Println("Devices found:")
	for _, device := range devices {

		fmt.Println("\nFlags: ", device.Flags)
		fmt.Println("\nName: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			fmt.Println("- IP address: ", address.IP)
			fmt.Println("- Subnet mask: ", address.Netmask)
		}
	}
}
