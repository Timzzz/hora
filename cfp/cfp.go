package cfp

import (
	"log"
	"time"

	"../adm"
	"../mondat"

	"github.com/spf13/viper"
)

type CfpController struct {
	cfps        map[string]Cfp
	m           adm.ADM
	monCh       chan mondat.TSPoint
	admCh       chan adm.ADM
	cfpResultCh chan Result
	lastNodeCpuPoint mondat.TSPoint
}

type Cfp interface {
	Insert(mondat.TSPoint)
	TSPoints() mondat.TSPoints
	Predict() (Result, error)
}

type Result struct {
	Component adm.Component
	Timestamp time.Time
	Predtime  time.Time
	PredMean  float64
	PredLB    float64
	PredUB    float64
	PredSd    float64
	FailProb  float64
}

func NewController(model adm.ADM) (*CfpController, <-chan Result) {
	viper.SetDefault("prediction.interval", "1m")
	viper.SetDefault("prediction.leadtime", "10m")
	viper.SetDefault("cfp.responsetime.unit", "1ns")
	viper.SetDefault("cfp.responsetime.threshold", "500ms")
	viper.SetDefault("cfp.responsetime.history", "20m")

	c := CfpController{
		cfps:        make(map[string]Cfp),
		m:           model,
		monCh:       make(chan mondat.TSPoint, 10),
		admCh:       make(chan adm.ADM, 1),
		cfpResultCh: make(chan Result, 10),
	}
	c.start()
	return &c, c.cfpResultCh
}

func (c *CfpController) AddMonDat(d mondat.TSPoint) {
	c.monCh <- d
}

func (c *CfpController) UpdateADM(m adm.ADM) {
	c.admCh <- m
}

func (c *CfpController) start() {
	log.Print("A Starting CfpController")
	go func() {
	Loop:
		for {
			log.Print("Cfp select...")
			select {
			case tsPoint, ok := <-c.monCh:
				if !ok {
					log.Print("cfp: Break CfpController")
					break Loop
				}
				comp := tsPoint.Component
				cfp, ok := c.cfps[comp.UniqName()]
				if !ok {
					log.Printf("cfp: predicting - comp type: %s", comp.Type)
					var err error
					// TODO: choose predictor based on component type
					interval := viper.GetDuration("prediction.interval")
					leadtime := viper.GetDuration("prediction.leadtime")
					switch comp.Type {
					case "responsetime":
						history := viper.GetDuration("cfp.responsetime.history")
						threshold := float64(viper.GetDuration("cfp.responsetime.threshold") / viper.GetDuration("cfp.responsetime.unit"))
						cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
						if err != nil {
							log.Printf("cfp: %s. %s", comp.UniqName(), err)
						}
						c.cfps[comp.UniqName()] = cfp
					case "cpu":
						history := viper.GetDuration("cfp.cpu.history")
						threshold := viper.GetFloat64("cfp.cpu.threshold")
						cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
						if err != nil {
							log.Printf("cfp: %s. %s", comp.UniqName(), err)
						}
						c.cfps[comp.UniqName()] = cfp
					case "nodecpu":
						history := viper.GetDuration("cfp.nodecpu.history")
						threshold := viper.GetFloat64("cfp.nodecpu.threshold")
						cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
						c.lastNodeCpuPoint = tsPoint
						if err != nil {
							log.Printf("cfp: %s. %s", comp.UniqName(), err)
						}
						c.cfps[comp.UniqName()] = cfp
					case "memory":
						history := viper.GetDuration("cfp.memory.history")
						threshold := viper.GetFloat64("cfp.memory.threshold")
						cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
						if err != nil {
							log.Printf("cfp: %s. %s", comp.UniqName(), err)
						}
						c.cfps[comp.UniqName()] = cfp
					case "service":
						history := viper.GetDuration("cfp.service.history")
						threshold := viper.GetFloat64("cfp.service.threshold")
						cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
						if err != nil {
							log.Printf("cfp: %s. %s", comp.UniqName(), err)
						}
						c.cfps[comp.UniqName()] = cfp
					default:
						log.Printf("cfp: unknown component type: %s. Skipping", comp.Type)
						continue Loop
					}
				} else {
					log.Print("cfp: not predicting.")
				}
				cfp.Insert(tsPoint)
				res, err := cfp.Predict()
				if err != nil {
					log.Printf("cfp: %s. %s", comp.UniqName(), err)
				}
				c.cfpResultCh <- res
			case model, _ := <-c.admCh:
			log.Print("cfp: assigning model")
				c.m = model
			}
		}
		close(c.cfpResultCh)
	log.Print("cfp: end.")
	}()
}
