package main // Intelligrator v1.0

import (
	"flag"
	"fmt"
	"time"

	"github.com/autogrow/go-jelly/ig"
)

// currentDay returns the day of the month as a integer
func currentDay() int {
	d := time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)
	return d.Day()
}

func main() {
	var cfgFile string

	flag.StringVar(&cfgFile, "c", "", "config for the intelligrator")
	flag.Parse()

	if cfgFile == "" {
		fmt.Println("Please specific a config file")
		return
	}

	cfg, err := newConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := ig.NewClient(cfg.Username, cfg.Password)
	if err != nil {
		fmt.Printf("Error getting IntelliClient: %s\n", err)
		return
	}
	fmt.Printf("IntelliClient created successful\n")

	err = c.GetDevices()

	if err != nil {
		fmt.Printf("Error getting devices: %s", err)
		return
	}
	fmt.Printf("Got valid devices for client\n")

	printStatus := time.NewTicker(time.Minute)
	tickTime := time.Duration(cfg.SampleTime)
	ticker := time.NewTicker(tickTime * time.Second)

	var accumLight float64
	var irrigationCounter int
	var light float64
	var valid bool

	// Convert the trigger level to J/cm2
	triggerLevel := cfg.TriggerLevel * 1000000.0
	day := currentDay()

	for {
		select {
		case <-ticker.C:
			light, valid = getLightReading(c, cfg.sourceType, cfg.sourceName)
			if !valid {
				fmt.Printf("Light reading from %s : %s is not valid\n", cfg.sourceType, cfg.sourceName)
				continue
			}

			accumLight += light * float64(cfg.SampleTime)

			// Check for midnight event
			newDay := currentDay()
			if newDay != day {
				irrigationCounter = 0
				if cfg.ResetMidnight {
					accumLight = 0
					fmt.Println("Midnight detected clear accumulation")
				}
			}

			day = newDay

			if accumLight > triggerLevel {
				irrigationCounter++
				currTime := time.Now().Format(time.RFC3339)
				fmt.Printf("Accumlated light level of %.4f J/m2 reached, Irrigation %d started at %s\n", accumLight/1000000, irrigationCounter, currTime)
				accumLight = 0
				triggerIrrigation(c, cfg.targetType, cfg.targetName)
			}
		case <-printStatus.C:
			fmt.Printf("Current Light: %.2f W/cm2, Accumulation: %.4f J/m2\n", light, accumLight/1000000)
		}
	}
}

func getLightReading(c *ig.Client, src, name string) (float64, bool) {
	if src == growroom {
		gr, exists := c.Growroom(name)
		if !exists {
			fmt.Printf("Growroom %s doesn't exist\n", name)
			return 0, false
		}
		err := gr.Update()
		if err != nil {
			fmt.Println(err)
		}
		r := gr.Climate
		return r.Light, true
	}

	dev, err := c.IntelliClimate(name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	dev.GetMetrics()
	return dev.Metrics.Light, true
}

func triggerIrrigation(c *ig.Client, targ, name string) {
	if targ == growroom {
		gr, exists := c.Growroom(name)
		if !exists {
			fmt.Printf("Growroom %s doesn't exist\n", name)
			return
		}

		devs, err := gr.IntelliDoses()
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, dev := range devs {
			err := dev.ForceIrrigation()
			if err != nil {
				fmt.Printf("setting intellidose %s forcing irrgation failed $%s \n", dev.ID, err)
				continue
			}
		}
		return
	}

	dev, err := c.IntelliDose(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = dev.ForceIrrigation()
	if err != nil {
		fmt.Printf("setting intellidose %s forcing irrgation failed $%s \n", dev.ID, err)
		return
	}
}
