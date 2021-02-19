package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

func main() {
	c := flag.String("c", "config.json", "config file")
	flag.Parse()

	buf, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Printf("read config error: %s", err)
		return
	}
	// default config
	conf := Config{
		HarvestInteval: 1,
		DumpInteval:    30,
		Discover:       true,
		RegistryPath:   "registry",
		RegistryTTL:    15 * 86400,
	}
	if err = json.Unmarshal(buf, &conf); err != nil {
		log.Printf("read config error: %s", err)
		return
	}
	if len(conf.Input) == 0 || conf.Output.IsEmpty() {
		log.Print("missing input or output config")
		return
	}
	harvester, err := NewHarvester(&conf)
	if err != nil {
		log.Printf("start error: %s", err)
		return
	}
	if err = harvester.Start(); err != nil {
		log.Printf("start error: %s", err)
	}
}
