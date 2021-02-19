package main

import "net"

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
	return w.conn.WriteTo(buf, w.addr)
}

// Close no-op.
func (w PacketWriter) Close() error {
	return nil
}
