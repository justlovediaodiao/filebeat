package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"regexp"
)

type dynamic struct {
	Type     string
	Settings json.RawMessage
}

func (d dynamic) IsEmpty() bool {
	return d.Type == "" || d.Settings == nil
}

// Config config.
type Config struct {
	Input          []string
	Filter         dynamic
	Output         dynamic
	HarvestInteval int64 `json:"harvest_inteval"`
	DumpInteval    int64 `json:"dump_inteval"`
	Discover       bool
	RegistryPath   string `json:"registry_path"`
	RegistryTTL    int64  `json:"registry_ttl"`
}

// Filter match a line, return whether matched.
type Filter func([]byte) bool

// OutputFactory is a factory of output.
type OutputFactory func() (io.WriteCloser, error)

func regexFilter(pattern string) (Filter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.Match, nil
}

// NewOutputFactory return a factory of output by output config.
func NewOutputFactory(output dynamic) (OutputFactory, error) {
	if output.Type == "udp" {
		var v struct{ Address string }
		if err := json.Unmarshal(output.Settings, &v); err != nil {
			return nil, err
		}
		addr, err := net.ResolveUDPAddr("udp", v.Address)
		if err != nil {
			return nil, err
		}
		return func() (io.WriteCloser, error) {
			return NewPacketWriter(addr)
		}, nil
	}
	return nil, fmt.Errorf("unsuuported output: %s", output.Type)
}

// NewFilter return a filter by filter config.
func NewFilter(filter dynamic) (Filter, error) {
	if filter.Type == "regex" {
		var v struct{ Pattern string }
		if err := json.Unmarshal(filter.Settings, &v); err != nil {
			return nil, err
		}
		return regexFilter(v.Pattern)
	}
	return nil, fmt.Errorf("unsuuported filter: %s", filter.Type)
}
