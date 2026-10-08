module IDIG4110/ingest-service

replace IDIG4110/shared => ../shared

go 1.27.1

require (
	IDIG4110/shared v0.0.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/jackc/pgx/v5 v5.11.0
	go.yaml.in/yaml/v4 v4.0.0-rc.6
)

require (
	github.com/golang-migrate/migrate/v4 v4.20.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.30 // indirect
	github.com/twmb/franz-go v1.22.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.14.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
