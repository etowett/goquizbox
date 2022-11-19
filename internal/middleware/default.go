package middleware

import (
	"net"
	"net/http"
	"strings"

	"goquizbox/internal/util"
	"goquizbox/internal/web/auth"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/pkg/logging"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

const requestIDHeaderKey = "X-Request-ID"
const userAgentHeaderKey = "User-Agent"

func DefaultMiddlewares(
	sessionAuthenticator auth.SessionAuthenticator,
) []gin.HandlerFunc {

	return []gin.HandlerFunc{
		compressor(),
		corsMiddleware(),

		// must go first
		setRequestId(),
		setupAppContext(sessionAuthenticator),
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Auth-Token, X-goquizbox-User-ID")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-CSRF-Token, Authorization, X-Requested-With, X-Auth-Token, X-goquizbox-User-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func compressor() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

func setRequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := util.GenerateUUID()
		ctx := ctxhelper.WithRequestId(c.Request.Context(), requestId)
		c.Request = c.Request.WithContext(ctx)
		c.Header(requestIDHeaderKey, requestId)
		c.Next()
	}
}

func setupAppContext(
	sessionAuthenticator auth.SessionAuthenticator,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("setupAppContext")

		ipAddress, err := getIP(c.Request)
		if err != nil {
			logger.Warnf("Unable to parse ipAddress %v from remote address: %v", c.Request.RemoteAddr, err)
		} else {
			ctx = ctxhelper.WithIpAddress(ctx, ipAddress)
		}

		userAgent := c.Request.Header.Get(userAgentHeaderKey)
		ctx = ctxhelper.WithUserAgent(ctx, userAgent)

		tokenInfo, err := sessionAuthenticator.TokenInfoFromRequest(c.Request)
		if err != nil {
			if err != auth.ErrTokenNotProvided {
				logger.Errorf("failed to get token from request: %v", err)
			}
		} else {
			ctx = ctxhelper.WithTokenInfo(ctx, tokenInfo)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func getIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-FORWARDED-FOR")
	forwarded := strings.Split(ips, ",")

	if ips != "" && len(forwarded) > 0 {
		return forwarded[0], nil
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	return host, err
}
