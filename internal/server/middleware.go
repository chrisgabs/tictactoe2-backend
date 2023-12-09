package server

import (
	"net/http"
)

// HandlePreFlight return true if request is preflight
func HandlePreFlight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		// Handle preflight request
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		//w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%v%v:%v", FrontendProtocol, FrontendAddress, FrontendPort))
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}
