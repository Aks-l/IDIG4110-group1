package server

import "net/http"

const (
	Version          = "v1"
	AuthPath         = " /" + Version + "/auth"
	AuthRegisterPath = http.MethodPost + AuthPath + "/register"
	AuthLoginPath    = http.MethodPost + AuthPath + "/login"
	AuthRefreshPath  = http.MethodPost + AuthPath + "/refresh"
	AuthLogoutPath   = http.MethodPost + AuthPath + "/logout"
)
