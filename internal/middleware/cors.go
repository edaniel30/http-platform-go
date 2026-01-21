package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration.
// This is a re-export to avoid forcing users to import the main package for middleware configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// IsWildcardOrigin checks if the origins list contains only the wildcard "*"
func IsWildcardOrigin(origins []string) bool {
	return len(origins) == 1 && origins[0] == "*"
}

// CORS creates a CORS middleware with the given configuration
// Important: According to the CORS specification, when using wildcard origin "*",
// credentials (cookies, HTTP auth) cannot be allowed. This middleware enforces
// this requirement by automatically setting AllowCredentials to false when "*" is used.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	if IsWildcardOrigin(cfg.AllowedOrigins) {
		config.AllowAllOrigins = true
		// CORS spec requirement: credentials cannot be used with wildcard origin
		// This is enforced regardless of cfg.AllowCredentials value
		config.AllowCredentials = false
	} else {
		config.AllowOrigins = cfg.AllowedOrigins
	}

	return cors.New(config)
}
