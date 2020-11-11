// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"encoding/binary"
	"fmt"

	"inet.af/netaddr"
)

// IP4 is an IPv4 address.
type IP4 uint32

// IPFromNetaddr converts a netaddr.IP to an IP. Panics if !ip.Is4.
func IP4FromNetaddr(ip netaddr.IP) IP4 {
	ipbytes := ip.As4()
	return IP4(binary.BigEndian.Uint32(ipbytes[:]))
}

// Netaddr converts an IP to a netaddr.IP.
func (ip IP4) Netaddr() netaddr.IP {
	return netaddr.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func (ip IP4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func (ip IP4) IsMulticast() bool {
	return byte(ip>>24)&0xf0 == 0xe0
}

func (ip IP4) IsLinkLocalUnicast() bool {
	return byte(ip>>24) == 169 && byte(ip>>16) == 254
}

// IPHeader represents an IP packet header.
type IP4Header struct {
	IPProto IPProto
	IPID    uint16
	SrcIP   IP4
	DstIP   IP4
}

const ip4HeaderLength = 20

func (IP4Header) Len() int {
	return ip4HeaderLength
}

func (h IP4Header) Marshal(buf []byte) error {
	if len(buf) < ip4HeaderLength {
		return errSmallBuffer
	}
	if len(buf) > maxPacketLength {
		return errLargePacket
	}

	buf[0] = 0x40 | (ip4HeaderLength >> 2) // IPv4
	buf[1] = 0x00                          // DHCP, ECN
	binary.BigEndian.PutUint16(buf[2:4], uint16(len(buf)))
	binary.BigEndian.PutUint16(buf[4:6], h.IPID)
	binary.BigEndian.PutUint16(buf[6:8], 0) // flags, offset
	buf[8] = 64                             // TTL
	buf[9] = uint8(h.IPProto)
	binary.BigEndian.PutUint16(buf[10:12], 0) // blank IP header checksum
	binary.BigEndian.PutUint32(buf[12:16], uint32(h.SrcIP))
	binary.BigEndian.PutUint32(buf[16:20], uint32(h.DstIP))

	binary.BigEndian.PutUint16(buf[10:12], ipChecksum(buf[0:20]))

	return nil
}

// MarshalPseudo serializes the header into buf in the "pseudo-header"
// form required when calculating UDP checksums. Overwrites the first
// h.Length() bytes of buf.
func (h IP4Header) MarshalPseudo(buf []byte) error {
	if len(buf) < ip4HeaderLength {
		return errSmallBuffer
	}
	if len(buf) > maxPacketLength {
		return errLargePacket
	}

	length := len(buf) - ip4HeaderLength
	binary.BigEndian.PutUint32(buf[8:12], uint32(h.SrcIP))
	binary.BigEndian.PutUint32(buf[12:16], uint32(h.DstIP))
	buf[16] = 0x0
	buf[17] = uint8(h.IPProto)
	binary.BigEndian.PutUint16(buf[18:20], uint16(length))
	return nil
}

// ToResponse implements Header.
func (h *IP4Header) ToResponse() {
	h.SrcIP, h.DstIP = h.DstIP, h.SrcIP
	// Flip the bits in the IPID. If incoming IPIDs are distinct, so are these.
	h.IPID = ^h.IPID
}
