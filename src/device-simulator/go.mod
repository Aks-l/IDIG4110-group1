module IDIG4110/device-simulator

replace IDIG4110/shared => ../shared

go 1.27.1

require (
	IDIG4110/shared v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.5.1
	go.yaml.in/yaml/v4 v4.0.0-rc.6
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)
