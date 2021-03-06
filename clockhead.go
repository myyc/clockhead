package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var ROOT = "/sys/devices/system/cpu"

const (
	AC        = 0
	Battery   = 1
	Locked    = 2
	Undefined = 3
)

func pathExists(path string) bool {
	_, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func getCores() int {
	c, _ := cpu.Counts(true)
	return c
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func setValue(fn string, v string) {
	if err := os.WriteFile(fn, []byte(v), 0600); err != nil {
		log.Fatal(err)
	}
}

func setValueForCore(core int, k string, v string) {
	setValue(fmt.Sprintf("%s/cpu%d/cpufreq/%s", ROOT, core, k), v)
}

func getValue(fn string) string {
	content, err := os.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Trim(string(content[:]), "\n")
}

func getValueForCore(core int, k string) string {
	return getValue(fmt.Sprintf("%s/cpu%d/cpufreq/%s", ROOT, core, k))
}

func getGovernors() []string {
	s := getValueForCore(0, "scaling_available_governors")
	return strings.Split(s, " ")
}

func isValidGovernor(g string) bool {
	for _, s := range getGovernors() {
		if s == g {
			return true
		}
	}

	return false
}

func getGovernor(core int) string {
	return getValueForCore(core, "scaling_governor")
}

func setGovernor(core int, g string) {
	if isValidGovernor(g) {
		setValueForCore(core, "scaling_governor", g)
	}
}

func setFrequency(core int, f int) {
	sf := fmt.Sprintf("%d", f)
	setValueForCore(core, "scaling_setspeed", sf)
}

/*
	for some reason setting an unavailable frequency has unpredictable
	behaviour depending on the target. on my machine (quite old) it
	works fine except for the 1.7-2GHz band, and the clock often gets
	stuck at 2GHz at low load.
*/
func setClosestFrequency(core int, f int) {
	freqsl := strings.Split(getValueForCore(0, "scaling_available_frequencies"), " ")
	freq := f
	diff := 1e15
	for _, fs := range freqsl {
		nf, _ := strconv.Atoi(fs)
		if ndiff := math.Abs(float64(nf - f)); ndiff < diff {
			diff = ndiff
			freq = nf
		}
	}
	sf := fmt.Sprintf("%d", freq)

	setValueForCore(core, "scaling_setspeed", sf)
}

func getFrequency(core int) int {
	return parseInt(getValueForCore(core, "scaling_cur_freq"))
}

func setAllGovernors(g string) {
	if isValidGovernor(g) && getGovernor(0) != g {
		for c := 0; c < getCores(); c++ {
			setGovernor(c, g)
		}
	}
}

func getMaxFrequency() int {
	return parseInt(getValueForCore(0, "scaling_max_freq"))
}

func getMinFrequency() int {
	return parseInt(getValueForCore(0, "scaling_min_freq"))
}

func isOnAC() bool {
	acs := [2]string{"ADP1", "ADP1"}
	for _, ac := range acs {
		path := fmt.Sprintf("/sys/class/power_supply/%s/online", ac)
		if pathExists(path) {
			return getValue(path) == "1"
		}
	}

	log.Fatal("Can't find AC path in /sys/class/power_supply, please report this")
	return true
}

func isLocked() bool {
	_, err := os.Stat("/tmp/clockhead.lock")
	return !errors.Is(err, os.ErrNotExist)
}

type Summary struct {
	freq int
	perc float64
	chg  string
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Add debugging log")
	flag.Usage = func() {
		fmt.Println("clockhead [--debug]")
	}
	flag.Parse()

	minf := getMinFrequency()
	maxf := getMaxFrequency()
	step := 250000
	interval := 3

	state := Undefined

	for {
		if isLocked() {
			if state != Locked {
				state = Locked
				println("Locked. Waiting ...")
			}
			time.Sleep(time.Duration(interval) * time.Second)
		} else {
			if state == Locked {
				println("Unlocked")
			}
			if isOnAC() {
				if state != AC {
					state = AC
					println("On AC ???????. Setting `performance` governor.")
					setAllGovernors("performance")
				}
				time.Sleep(time.Duration(interval) * time.Second)
			} else {
				if state != Battery {
					state = Battery
					println("On Battery ???????. Setting `userspace` governor and optimising frequency.")
					setAllGovernors("userspace")
				}
				percs, _ := cpu.Percent(time.Duration(interval)*time.Second, true)

				summary := make([]Summary, getCores())

				for core, perc := range percs {
					freq := getFrequency(core)
					summary[core].chg = ""

					if perc > 90 {
						if freq+3*step < maxf {
							setClosestFrequency(core, freq+3*step)
							summary[core].chg = "??????  ?????? "
						} else {
							setFrequency(core, maxf)
							summary[core].chg = "????"
						}
					} else if perc > 50 {
						if freq+step < maxf {
							setClosestFrequency(core, freq+step)
							summary[core].chg = "?????? "
						} else {
							setFrequency(core, maxf)
							summary[core].chg = "????"
						}
					} else if perc < 3 {
						if freq-2*step > minf {
							setClosestFrequency(core, freq-2*step)
							summary[core].chg = "?????? ??????"
						} else {
							setFrequency(core, minf)
						}
					} else if perc < 10 {
						if freq-step > minf {
							setClosestFrequency(core, freq-step)
							summary[core].chg = "??????"
						} else {
							setFrequency(core, minf)
						}
					}

					summary[core].perc = perc
					summary[core].freq = getFrequency(core)
				}

				if debug {
					for core, s := range summary {
						str := fmt.Sprintf("%d:\t%.2f%%, %.2fGHz", core, s.perc, float64(s.freq)/1e6)
						if s.chg != "" {
							str = fmt.Sprintf("%s %s", str, s.chg)
						}
						println(str)
					}
					println("")
				}
			}
		}
	}
}
