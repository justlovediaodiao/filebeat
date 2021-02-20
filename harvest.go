package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Harvester read from input and write to output.
type Harvester struct {
	registry  Registry
	inputs    StringSet
	notify    chan FileOffset
	outputFac OutputFactory
	filter    Filter
	config    *Config
}

// NewHarvester return a Harvester.
func NewHarvester(config *Config) (*Harvester, error) {
	outputFac, err := NewOutputFactory(config.Output)
	if err != nil {
		return nil, err
	}
	var filter Filter
	if !config.Filter.IsEmpty() {
		filter, err = NewFilter(config.Filter)
		if err != nil {
			return nil, err
		}
	}
	return &Harvester{
		registry:  NewRegistry(config.RegistryPath, config.RegistryTTL),
		inputs:    make(StringSet),
		notify:    make(chan FileOffset, 32),
		outputFac: outputFac,
		filter:    filter,
		config:    config,
	}, nil
}

// Start start harvest.
func (h *Harvester) Start() error {
	if err := h.registry.Load(); err != nil {
		return err
	}
	// scan input.
	for path := range h.scan() {
		if err := h.startNewBeater(path); err != nil {
			log.Printf("start new beater %s error: %s", path, err)
		}
	}
	h.serve()
	return nil
}

func (h *Harvester) startNewBeater(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	id := FileUID(info)
	of, ok := h.registry.Get(id)
	var offset int64
	if ok {
		if of.Offset == info.Size() { // harvest ended last time.
			return nil
		}
		if of.Offset < info.Size() { // truncated, harvest from head.
			offset = of.Offset
		}
	}
	beater, err := NewFileBeater(id, path, offset, time.Duration(h.config.HarvestInteval*int64(time.Second)))
	if err != nil {
		return err
	}
	output, err := h.outputFac()
	if err != nil {
		beater.Close()
		return err
	}
	h.inputs.Add(beater.ID)
	go h.harvest(beater, output)
	log.Printf("start harvesting %s from offset %d", path, offset)
	return nil
}

func (h *Harvester) serve() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	interval := time.Duration(h.config.DumpInteval * int64(time.Second))
	t := time.NewTimer(interval)
	for {
		select {
		case v := <-h.notify:
			if v.Offset == -1 { // harvest ended
				h.inputs.Remove(v.ID)
			} else {
				h.registry.Set(v)
			}
		case <-t.C: // dump registry and rescan input.
			if err := h.registry.Dump(); err != nil {
				log.Printf("dump offset error: %s", err)
			}
			if h.config.Discover {
				for path := range h.scan() {
					info, err := os.Stat(path)
					if err != nil {
						continue
					}
					id := FileUID(info)
					if !h.inputs.Has(id) {
						if err := h.startNewBeater(path); err != nil {
							log.Printf("start new beater %s error: %s", path, err)
						}
					}
				}
			}
			t.Reset(interval)
		case <-sig:
			h.registry.Dump()
			return
		}
	}
}

func (h *Harvester) scan() StringSet {
	result := make(StringSet)
	for _, glob := range h.config.Input {
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Printf("match glob %s failed: %s", glob, err)
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err == nil && !info.IsDir() {
				result.Add(match)
			}
		}
	}
	return result
}

func (h *Harvester) harvest(input *FileBeater, output io.WriteCloser) {
	defer input.Close()
	defer output.Close()
	for {
		buf, err := input.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Printf("%s harvest finished", input.Path)
			} else {
				log.Printf("read %s error: %s", input.Path, err)
			}
			break
		}
		if h.filter == nil || h.filter(buf) {
			_, err = output.Write(buf)
			if err != nil { // should not write again if a writer error.
				log.Printf("write error: %s", err)
				break
			}
		}
		h.notify <- FileOffset{input.ID, input.Offset, time.Now().Unix()}
	}
	h.notify <- FileOffset{input.ID, -1, time.Now().Unix()}
}
