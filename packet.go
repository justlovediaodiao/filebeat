package main

import "net"

// maxPacketSize is the max size of a udp packet on linux.
// 2-bytes - ip header size - udp header size
const maxPacketSize = 65535 - 8 - 20

// PacketWriter write to udp.
type PacketWriter struct {
	conn net.PacketConn
	addr net.Addr
}

// NewPacketWriter return a PacketWriter with target address.
func NewPacketWriter(addr net.Addr) (*PacketWriter, error) {
	conn, err := net.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}
	return &PacketWriter{conn, addr}, nil
}

// Write write data.
func (w PacketWriter) Write(buf []byte) (int, error) {
	if len(buf) > maxPacketSize {
		buf = buf[:maxPacketSize]
	}
	return w.conn.WriteTo(buf, w.addr)
}

// Close no-op.
func (w PacketWriter) Close() error {
	return nil
}
