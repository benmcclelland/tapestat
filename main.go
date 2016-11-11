package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"unicode/utf8"
)

type unit struct {
	unit    string
	divisor float64
}

func (u *unit) String() string {
	return fmt.Sprintf("%v", u.unit)
}

func (u *unit) Set(s string) error {
	if s == "" {
		s = "B"
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return fmt.Errorf("unable to decode rune from %v", s)
	}
	switch r {
	case 'B':
		fallthrough
	case 'b':
		u.divisor = 1
		u.unit = " B"
	case 'K':
		fallthrough
	case 'k':
		u.divisor = 1024
		u.unit = "KB"
	case 'M':
		fallthrough
	case 'm':
		u.divisor = 1048576
		u.unit = "MB"
	case 'G':
		fallthrough
	case 'g':
		u.divisor = 1073741824
		u.unit = "GB"
	case 'T':
		fallthrough
	case 't':
		u.divisor = 1099511627776
		u.unit = "TB"
	default:
		return fmt.Errorf("invalid unit")

	}
	return nil
}

func UnitVar(p *unit, name string, value string, usage string) {
	err := p.Set(value)
	if err != nil {
		log.Fatalln("Bad default value (%v): %v", value, err)
	}
	flag.Var(p, name, usage)
}

type Config struct {
	unit       unit
	continuous bool
}

var config Config

func getDevice(dev string) Stats {
	stat, err := GetStats(dev)
	if err != nil {
		log.Fatalln(err)
	}
	return stat
}

func main() {
	UnitVar(&config.unit, "u", "B", "display units")
	flag.BoolVar(&config.continuous, "c", false, "continuous display per second")
	flag.Parse()

	devs, err := FindDevices()
	if err != nil {
		log.Fatalln(err)
	}

	devstats := map[string]Stats{}

	for _, dev := range devs {
		devstats[dev] = getDevice(dev)
	}
	time.Sleep(1 * time.Second)
	newstats := map[string]Stats{}
	if config.continuous {
		for {
			fmt.Printf("Dev:\t   r%s/s     r/s    w%s/s     w/s  other/s    %%util\n", config.unit.String(), config.unit.String())
			for _, dev := range devs {
				newstats[dev] = getDevice(dev)
				fmt.Printf("%v\t", dev)

				// read
				diff := newstats[dev]["read_byte_cnt"] - devstats[dev]["read_byte_cnt"]
				fmt.Printf("%8.2f", float64(diff)/config.unit.divisor)
				devstats[dev]["read_byte_cnt"] = newstats[dev]["read_byte_cnt"]
				diff = newstats[dev]["read_cnt"] - devstats[dev]["read_cnt"]
				fmt.Printf("%8.2f ", float64(diff))
				devstats[dev]["read_cnt"] = newstats[dev]["read_cnt"]

				// write
				diff = newstats[dev]["write_byte_cnt"] - devstats[dev]["write_byte_cnt"]
				fmt.Printf("%8.2f", float64(diff)/config.unit.divisor)
				devstats[dev]["write_byte_cnt"] = newstats[dev]["write_byte_cnt"]
				diff = newstats[dev]["write_cnt"] - devstats[dev]["write_cnt"]
				fmt.Printf("%8.2f ", float64(diff))
				devstats[dev]["write_cnt"] = newstats[dev]["write_cnt"]

				// other
				diff = newstats[dev]["other_cnt"] - devstats[dev]["other_cnt"]
				fmt.Printf("%8.2f ", float64(diff))
				devstats[dev]["other_cnt"] = newstats[dev]["other_cnt"]

				// util
				diff = newstats[dev]["io_ns"] - devstats[dev]["io_ns"]
				fmt.Printf("%8.2f ", float64(diff)/10000000)
				devstats[dev]["io_ns"] = newstats[dev]["io_ns"]

				fmt.Println()
			}
			fmt.Println()
			time.Sleep(1 * time.Second)
		}
	} else {
		fmt.Printf("Dev:\t     r%s   reads      w%s  writes  other/s\n", config.unit.String(), config.unit.String())

		for _, dev := range devs {
			newstats[dev] = getDevice(dev)
			fmt.Printf("%v\t", dev)
			fmt.Printf("%8.2f", float64(newstats[dev]["read_byte_cnt"])/config.unit.divisor)
			fmt.Printf("%8.2f ", float64(newstats[dev]["read_cnt"]))
			fmt.Printf("%8.2f", float64(newstats[dev]["write_byte_cnt"])/config.unit.divisor)
			fmt.Printf("%8.2f ", float64(newstats[dev]["write_cnt"]))
			fmt.Printf("%8.2f ", float64(newstats[dev]["other_cnt"]))
			fmt.Println()
		}
	}
}
