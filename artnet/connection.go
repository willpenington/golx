package artnet

import (
  "net"
  "bytes"
)

const artnetPort int = 6465

var bconn *net.UDPConn
var lconn *net.UDPConn

var bcastAddr *net.UDPAddr
var localAddr *net.UDPAddr

var localIP net.IP

func init() {
  localIP = findLocalIP()
  bcastIP := findBroadcastIP(localIP)

  bcastAddr = new(net.UDPAddr)
  bcastAddr.IP = bcastIP
  bcastAddr.Port = artnetPort
  localAddr, _ = net.ResolveUDPAddr("udp", ":" + string(artnetPort))

  var err error

  bconn, err = net.ListenUDP("udp", bcastAddr)
  if err == nil {
    go listen(bconn)
  }

  lconn, err = net.ListenUDP("udp", localAddr)
  if err == nil {
    go listen(lconn)
  }
}

func findLocalIP() net.IP {
  var secondaryIP net.IP = nil

  // Try and find an IP in 2.0.0.0/8 and if one can't be found an IP in
  // 10.0.0.0/8

  interfaces , _ := net.Interfaces()
  for _, ifi := range interfaces {
    addrs, _ := ifi.Addrs()

    for _, addr := range addrs {
      // Less hacky than reflection...
      ipAddr, _, err := net.ParseCIDR(addr.String())

      if err == nil {
        if ipAddr[0] == 2 {
          return ipAddr
        } else if ipAddr[0] == 10 {
          secondaryIP = ipAddr
        }
      }
    }
  }

  // If nothing was found in 2.0.0.0/8 but there was a 10.0.0.0/8 address use that
  if secondaryIP != nil {
    return secondaryIP
  }

  // If there is still nothing pick the first non-loopback IP the system has
  addrs, _ := net.InterfaceAddrs()
  for _, addr := range addrs {
    ipAddr, _, err := net.ParseCIDR(addr.String())

    if err == nil {
      if !ipAddr.IsLoopback() {
        return ipAddr
      }
    }
  }

  // If there truly is nothing but loopback, set localIP to localhost
  loopback := net.ParseIP("127.0.0.1")
  return loopback
}

// Find the correct Artnet broadcast address
func findBroadcastIP(local net.IP) net.IP {
  if local[0] == 10 {
    return net.ParseIP("10.255.255.255")
  } else {
    return net.ParseIP("2.255.255.255")
  }
}

// Read incoming requests and dispatch packets
func listen(conn *net.UDPConn) {
  for {
    data := make([]byte, 4096)
    n, addr, err := conn.ReadFromUDP(data)
    if (n > 0) && (err == nil) {
      buffer := bytes.NewBuffer(data)
      go HandlePacket(buffer, addr)
    }
  }
}

// Send a packet to a given address
func sendPacket(data []byte, addr *net.UDPAddr) {
  lconn.WriteToUDP(data, addr)
}

// Send a packet to every Artnet node
func broadcast(data []byte) {
  sendPacket(data, bcastAddr)
}
