package middleware

import (
	"net/http"
	"strings"
)

var allowedHeaders = strings.Join([]string{
	"Accept",
	"Authorization",
	"Content-Type",
	"Origin",
	"Referer",
	"User-Agent",
	"device-generated-id",
	"platform-id",
	"sec-ch-ua",
	"sec-ch-ua-mobile",
	"sec-ch-ua-platform",
}, ", ")

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, HEAD, PUT, PATCH, DELETE")
			resp.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			resp.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if req.Method == http.MethodOptions {
			resp.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(resp, req)
	})
}
