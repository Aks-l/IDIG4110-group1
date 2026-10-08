package server

const (
	VERSION   = "v1"
	API_ROUTE = "/api/" + VERSION
	
	HEALTHZ = "/healthz"

	INGEST_ROUTE = API_ROUTE + "/ingest"

	SENSOR_DATA = INGEST_ROUTE + "/sensor-data"

	SENSORS = INGEST_ROUTE + "/sensors"
)
