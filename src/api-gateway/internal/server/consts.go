package server

const (
	VERSION   = "v1"
	API_ROUTE = "/api/" + VERSION

	GATEWAY_ROUTE = API_ROUTE + "/gateway"

	HEALTH = GATEWAY_ROUTE + "/health"
)
