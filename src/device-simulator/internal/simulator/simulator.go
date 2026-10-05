package simulator

import (
	"IDIG4110/device-simulator/internal/broker"
)

type Simulator struct {
	broker      *broker.Broker
	interval    int
	topic       string
	deviceCount int
	qos         byte
}

func Init(b broker.Broker, ) *Simulator {
	return nil
}
