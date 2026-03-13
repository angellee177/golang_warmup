package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	// Helper function to create a token for testing
	createToken := func(sub string, expired bool) string {
		exp := time.Now().Add(time.Hour).Unix()
		if expired {
			exp = time.Now().Add(-time.Hour).Unix()
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": sub,
			"exp": exp,
		})
		s, _ := token.SignedString([]byte(secret))
		return s
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Token Format",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired Token",
			authHeader:     "Bearer " + createToken("123", true),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + createToken("user-uuid-123", false),
			expectedStatus: http.StatusOK,
			expectedUserID: "user-uuid-123",
		},
		{
			name: "Invalid Token Payload - sub is not a string",
			authHeader: "Bearer " + func() string {
				// Manually create a token where 'sub' is an integer instead of a string
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": 12345, // This will fail the .(string) type assertion
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				s, _ := token.SignedString([]byte(secret))
				return s
			}(),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid Signing Method - None instead of HMAC",
			authHeader: "Bearer " + func() string {
				// Create a token with the 'None' method
				token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
					"sub": "user-123",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				// Sign it using the special 'Unsafe' constant required for 'None'
				s, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return s
			}(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router and route
			r := gin.New()
			r.Use(AuthMiddleware(secret))
			r.GET("/test", func(c *gin.Context) {
				uid, _ := c.Get("userId")
				c.JSON(http.StatusOK, gin.H{"userId": uid})
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), tt.expectedUserID)
			}
		})
	}
}
