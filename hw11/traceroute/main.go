package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"
)

const DEFAULT_PORT = 33434
const DEFAULT_MAX_BOUNCES = 30
const DEFAULT_TIMEOUT_MS = 500
const DEFAULT_RETRIES = 3
const DEFAULT_PACKET_SIZE = 60

func main() {
	var max_bounces = flag.Int("max_bounces", DEFAULT_MAX_BOUNCES, `Set the max time-to-live`)
	flag.Parse()

	options := TracerouteOptions{
		port:       DEFAULT_PORT,
		retries:    DEFAULT_RETRIES,
		maxBounces: *max_bounces,
		timeoutMs:  DEFAULT_TIMEOUT_MS,
		packetSize: DEFAULT_PACKET_SIZE,
	}

	host := flag.Arg(0)
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.maxBounces, options.packetSize)

	var (
		c  = make(chan TracerouteBounce, *max_bounces)
		wg = sync.WaitGroup{}
	)
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		for bounce := range c {
			addr := fmt.Sprintf("%v.%v.%v.%v", bounce.Address[0], bounce.Address[1], bounce.Address[2], bounce.Address[3])
			hostOrAddr := addr
			if bounce.Host != "" {
				hostOrAddr = bounce.Host
			}
			if bounce.Success {
				fmt.Printf("%-3d %v (%v)  %v\n", bounce.TTL, hostOrAddr, addr, bounce.ElapsedTime)
			} else {
				fmt.Printf("%-3d *\n", bounce.TTL)
			}
		}
		fmt.Println()
	}()

	err = Traceroute(host, &options, c)
	wg.Wait()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
}

type ICMP struct {
	Type        byte
	Code        byte
	Checksum    uint16
	MessageSize int
	Message     []byte
}

func (i *ICMP) toBytes() []byte {
	data := make([]byte, 9+i.MessageSize)
	data[0] = i.Type
	data[1] = i.Code
	data[2] = uint8(i.Checksum >> 8)
	data[3] = uint8(i.Checksum & 0xff)
	copy(data[4:], i.Message)
	return data
}

func (i *ICMP) checksum() uint16 {
	var (
		acc  = uint32(0)
		data = i.toBytes()
	)

	for j := 0; j < i.MessageSize+8; j += 2 {
		acc += uint32(data[j])<<8 + uint32(data[j+1])
		j += 2
	}
	acc = (acc >> 16) + (acc & 0xffff)
	acc += (acc >> 16)
	return (uint16)(^acc)
}

func setUpICMP() []byte {
	var (
		data = []byte("ping")
		icmp = ICMP{Type: 0x08, Code: 0x00, MessageSize: len(data) + 4}
	)
	icmp.Message = make([]byte, icmp.MessageSize)
	icmp.Message[1] = 0x01
	icmp.Message[3] = 0x01
	copy(icmp.Message[4:], data)
	icmp.Checksum = icmp.checksum()
	return icmp.toBytes()
}

func socketAddr() (addr [4]byte, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if len(ipnet.IP.To4()) == net.IPv4len {
				copy(addr[:], ipnet.IP.To4())
				return
			}
		}
	}
	err = errors.New("You do not appear to be connected to the Internet")
	return
}

func destAddr(dest string) (destAddr [4]byte, err error) {
	addrs, err := net.LookupHost(dest)
	if err != nil {
		return
	}
	addr := addrs[0]

	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return
	}
	copy(destAddr[:], ipAddr.IP.To4())
	return
}

type TracerouteOptions struct {
	port       int
	maxBounces int
	timeoutMs  int
	retries    int
	packetSize int
}

type TracerouteBounce struct {
	Success     bool
	Address     [4]byte
	Host        string
	ElapsedTime time.Duration
	TTL         int
}

func toNanos(ms int) int64 {
	return 1000 * 1000 * int64(ms)
}

func Traceroute(dest string, options *TracerouteOptions, c chan TracerouteBounce) (err error) {
	defer close(c)

	destAddr, err := destAddr(dest)
	if err != nil {
		return
	}
	socketAddr, err := socketAddr()
	if err != nil {
		return
	}

	tv := syscall.NsecToTimeval(toNanos(options.timeoutMs))

	recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return err
	}
	defer syscall.Close(recvSocket)

	sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_ICMP)
	if err != nil {
		return err
	}
	defer syscall.Close(sendSocket)

	syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.port, Addr: socketAddr})
	for ttl := 1; ttl <= options.maxBounces; ttl++ {
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl)
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		flag := false
		for retry := 0; retry < options.retries; retry++ {
			start := time.Now()

			syscall.Sendto(sendSocket, setUpICMP(), 0, &syscall.SockaddrInet4{Port: options.port, Addr: destAddr})
			var p = make([]byte, options.packetSize)
			_, from, err := syscall.Recvfrom(recvSocket, p, 0)

			elapsed := time.Since(start)

			if err == nil {
				currAddr := from.(*syscall.SockaddrInet4).Addr

				bounce := TracerouteBounce{Success: true, Address: currAddr, ElapsedTime: elapsed, TTL: ttl}

				currHost, err := net.LookupAddr(fmt.Sprintf("%v.%v.%v.%v", bounce.Address[0], bounce.Address[1], bounce.Address[2], bounce.Address[3]))
				if err == nil {
					bounce.Host = currHost[0]
				} else {
					bounce.Host = "unknown"
				}

				c <- bounce

				if currAddr == destAddr {
					return nil
				}

				flag = true
				break
			}

			if ttl > options.maxBounces {
				return nil
			}
		}

		if !flag {
			c <- TracerouteBounce{Success: false, TTL: ttl}
		}
	}
	return nil
}
