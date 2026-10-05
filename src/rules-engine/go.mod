module IDIG4110/rules-engine

go 1.27.1

require IDIG4110/shared v0.0.0

require (
	github.com/golang-migrate/migrate/v4 v4.20.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
)

replace IDIG4110/shared => ../shared
