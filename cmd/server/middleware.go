package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/snowpackdata/cronos"
)

// AppContextKey is used as the key for storing and retrieving the App from the context
type AppContextKey string

// AppContextMiddleware adds the App instance to the request context
func (a *App) AppContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new context with the App instance
		ctx := context.WithValue(r.Context(), AppContextKey("app"), a)
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Claims is a non-persistent object that is used to store the JWT token and associated information
type Claims struct {
	UserID           uint
	TenantID         uint
	AccountID        uint
	Email            string
	IsStaff          bool
	Role             string // Add specific role (ADMIN, STAFF, CLIENT)
	RegisteredClaims *jwt.RegisteredClaims
}

func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.RegisteredClaims == nil {
		return nil, nil
	}
	return c.RegisteredClaims.ExpiresAt, nil
}

func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.RegisteredClaims == nil {
		return nil, nil
	}
	return c.RegisteredClaims.IssuedAt, nil
}

func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	if c.RegisteredClaims == nil {
		return nil, nil
	}
	return c.RegisteredClaims.NotBefore, nil
}

func (c Claims) GetIssuer() (string, error) {
	if c.RegisteredClaims == nil {
		return "", nil
	}
	return c.RegisteredClaims.Issuer, nil
}

func (c Claims) GetSubject() (string, error) {
	return c.Email, nil
}

func (c Claims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.Audience, nil
}

// IsValidIssuer checks if the token issuer is from a trusted source
func IsValidIssuer(issuer string) bool {
	// Accept tokens from localhost or snowpackdata.com
	validIssuers := []string{"localhost", "snowpackdata.com", ""}
	for _, valid := range validIssuers {
		if issuer == valid {
			return true
		}
	}
	return false
}

func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		next.ServeHTTP(w, r)
	})
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// Exception sends messages to frontend via json
type Exception struct {
	Message string `json:"message"`
}

// ParseTokenAndSetUserContext attempts to parse a JWT from the x-access-token header.
// If the token is valid, it populates the request context with user_id, account_id, user_email, and is_staff.
// It does NOT block requests if the token is missing or invalid.
func ParseTokenAndSetUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		tokenString := r.Header.Get("x-access-token")
		tokenString = strings.TrimSpace(tokenString)

		if tokenString == "" {
			log.Println("ParseTokenAndSetUserContext: No token found in header for page load.")
			next.ServeHTTP(w, r) // Proceed without user context
			return
		}

		tclaims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, tclaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})

		issuer := ""
		if tclaims.RegisteredClaims != nil {
			issuer = tclaims.RegisteredClaims.Issuer
		}

		if err == nil && token.Valid && IsValidIssuer(issuer) {
			log.Printf("ParseTokenAndSetUserContext: Token valid. UserID: %d, TenantID: %d, AccountID: %d, Email: %s, IsStaff: %v, Role: %s",
				tclaims.UserID, tclaims.TenantID, tclaims.AccountID, tclaims.Email, tclaims.IsStaff, tclaims.Role)
			ctx := context.WithValue(r.Context(), "user_id", tclaims.UserID)
			ctx = context.WithValue(ctx, "TenantId", tclaims.TenantID)
			ctx = context.WithValue(ctx, "account_id", tclaims.AccountID)
			ctx = context.WithValue(ctx, "user_email", tclaims.Email)
			ctx = context.WithValue(ctx, "is_staff", tclaims.IsStaff)
			ctx = context.WithValue(ctx, "user_role", tclaims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if err != nil {
			log.Printf("ParseTokenAndSetUserContext: Token parsing error: %v", err)
		} else if !token.Valid {
			log.Printf("ParseTokenAndSetUserContext: Token marked invalid.")
		} else if !IsValidIssuer(issuer) {
			log.Printf("ParseTokenAndSetUserContext: Invalid issuer: %s", issuer)
		}

		log.Println("ParseTokenAndSetUserContext: Token invalid or not present, proceeding without setting user context.")
		next.ServeHTTP(w, r)
	})
}

// RequireStaff checks if a user is authenticated and is staff.
// Expects ParseTokenAndSetUserContext to have run first.
func RequireStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isStaff, ok := r.Context().Value("is_staff").(bool)
		userID, userOk := r.Context().Value("user_id").(uint)

		if !ok || !userOk || userID == 0 || !isStaff {
			log.Printf("RequireStaff: Access denied. UserID: %v, IsStaff: %v. Redirecting to /404.", r.Context().Value("user_id"), r.Context().Value("is_staff"))
			http.Redirect(w, r, "/404", http.StatusFound)
			return
		}
		log.Printf("RequireStaff: Access granted for staff UserID: %d to %s", userID, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// RequireValidUser checks if a user is authenticated (has a valid user_id in context).
// Expects ParseTokenAndSetUserContext to have run first.
func RequireValidUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("user_id").(uint)
		userEmail, _ := r.Context().Value("user_email").(string)

		if !ok || userID == 0 {
			log.Printf("RequireValidUser: Access denied. No valid user in context. Redirecting to /login for path %s", r.URL.Path)
			http.Redirect(w, r, "/login", http.StatusFound) // Or a more generic error/page
			return
		}
		log.Printf("RequireValidUser: Access granted for UserID: %d (%s) to %s", userID, userEmail, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// JwtVerify API authentication: Verifies token and blocks if invalid/missing.
// This is the original middleware, primarily for API routes.
// It can be simplified if ParseTokenAndSetUserContext handles dev mode bypasses.
// For now, keeping its original structure largely, but it will use context if already populated.
func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("JwtVerify (API): Processing request for %s", r.URL.Path)

		// Check if user context is already populated by ParseTokenAndSetUserContext
		userIDCtx, userIDCtxOk := r.Context().Value("user_id").(uint)
		isStaffCtx, isStaffCtxOk := r.Context().Value("is_staff").(bool)
		roleCtx, roleCtxOk := r.Context().Value("user_role").(string)
		// We can also check for account_id here if JwtVerify needs to be account-aware for some APIs
		// accountIDCtx, accountIDCtxOk := r.Context().Value("account_id").(uint)

		if userIDCtxOk && userIDCtx > 0 && isStaffCtxOk && roleCtxOk { // isStaffCtxOk ensures it was explicitly set
			log.Printf("JwtVerify (API): User context already populated. UserID: %d, IsStaff: %v, Role: %s. Allowing.", userIDCtx, isStaffCtx, roleCtx)
			next.ServeHTTP(w, r)
			return
		}

		// If context not populated, proceed with original JwtVerify logic for token extraction and validation for APIs
		appInstance, _ := r.Context().Value(AppContextKey("app")).(*App)
		tokenString := r.Header.Get("x-access-token")
		tokenString = strings.TrimSpace(tokenString)

		isLocalEnv := os.Getenv("ENVIRONMENT") == "local"

		// Use DevToken only in local environment when no token is present
		// Note: DevToken usage is currently commented out in main.go, so this path might not be hit
		if isLocalEnv && appInstance != nil && appInstance.DevToken != "" && tokenString == "" {
			log.Printf("JwtVerify (API): Using development JWT token (local env)")
			tokenString = appInstance.DevToken
		}

		if tokenString == "" {
			log.Println("JwtVerify (API): No token found, returning 403")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(Exception{Message: "Missing auth token"})
			return
		}

		tclaims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, tclaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})

		apiIssuer := ""
		if tclaims.RegisteredClaims != nil {
			apiIssuer = tclaims.RegisteredClaims.Issuer
		}

		if err != nil || !token.Valid || !IsValidIssuer(apiIssuer) {
			// Specific logic for local dev token to bypass full validation IF it's the DevToken
			// This part also needs to be aware that DevToken is currently disabled in main.go
			if isLocalEnv && appInstance != nil && tokenString == appInstance.DevToken {
				log.Printf("JwtVerify (API): Local dev mode with matching DevToken. Attempting to decode DevToken claims for API.")
				parts := strings.Split(tokenString, ".")
				if len(parts) == 3 {
					payload, decodeErr := base64.RawURLEncoding.DecodeString(parts[1])
					if decodeErr == nil {
						var devTokenClaims struct { // Using an anonymous struct for local dev token claim decoding
							UserID           uint                  `json:"UserID"`
							AccountID        uint                  `json:"AccountID"` // Expect AccountID in dev token too
							Email            string                `json:"Email"`
							IsStaff          bool                  `json:"IsStaff"`
							Role             string                `json:"Role"`
							RegisteredClaims *jwt.RegisteredClaims `json:"RegisteredClaims"`
						}
						if json.Unmarshal(payload, &devTokenClaims) == nil && devTokenClaims.RegisteredClaims != nil && IsValidIssuer(devTokenClaims.RegisteredClaims.Issuer) {
							log.Printf("JwtVerify (API): DevToken claims successfully decoded. UserID: %d, AccountID: %d, IsStaff: %v, Role: %s",
								devTokenClaims.UserID, devTokenClaims.AccountID, devTokenClaims.IsStaff, devTokenClaims.Role)
							ctx := context.WithValue(r.Context(), "user_id", devTokenClaims.UserID)
							ctx = context.WithValue(ctx, "account_id", devTokenClaims.AccountID)
							ctx = context.WithValue(ctx, "user_email", devTokenClaims.Email)
							ctx = context.WithValue(ctx, "is_staff", devTokenClaims.IsStaff)
							ctx = context.WithValue(ctx, "user_role", devTokenClaims.Role)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
				log.Printf("JwtVerify (API): DevToken found but could not be properly decoded or claims are invalid. Returning 403.")
			} // else, it's a regular token that failed validation

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			msg := "Invalid auth token"
			if err != nil {
				msg = fmt.Sprintf("Invalid auth token: %s", err.Error())
			}
			json.NewEncoder(w).Encode(Exception{Message: msg})
			return
		}

		// If we are here, it means a regular token was parsed successfully by ParseWithClaims earlier
		log.Printf("JwtVerify (API): Token validated successfully (standard path). UserID: %d, AccountID: %d, IsStaff: %v, Role: %s",
			tclaims.UserID, tclaims.AccountID, tclaims.IsStaff, tclaims.Role)
		ctx := context.WithValue(r.Context(), "user_id", tclaims.UserID)
		ctx = context.WithValue(ctx, "account_id", tclaims.AccountID)
		ctx = context.WithValue(ctx, "user_email", tclaims.Email)
		ctx = context.WithValue(ctx, "is_staff", tclaims.IsStaff)
		ctx = context.WithValue(ctx, "user_role", tclaims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantContextKey is used to store tenant in context
type TenantContextKey string

const TenantKey TenantContextKey = "tenant"

// TenantMiddleware extracts subdomain and loads tenant from database
func TenantMiddleware(app *cronos.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := extractSubdomain(r.Host)

			// Handle reserved/invalid subdomains - allow empty for login page
			if slug == "" || slug == "www" || slug == "app" {
				// Allow through for login pages, they'll redirect to tenant subdomain
				next.ServeHTTP(w, r)
				return
			}

			reserved := map[string]bool{
				"api": true, "admin": true, "www": true,
				"mail": true, "ftp": true, "app": true,
			}
			if reserved[slug] {
				http.Error(w, "Invalid tenant", http.StatusBadRequest)
				return
			}

			// Load tenant from database
			var tenant cronos.Tenant
			if err := app.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
				log.Printf("TenantMiddleware: Tenant not found for slug '%s': %v", slug, err)
				http.Error(w, "Tenant not found", http.StatusNotFound)
				return
			}

			log.Printf("TenantMiddleware: Loaded tenant: %s (ID: %d)", tenant.Name, tenant.ID)

			// Add tenant to context
			ctx := context.WithValue(r.Context(), TenantKey, &tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractSubdomain extracts the subdomain from a host string
func extractSubdomain(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")

	// Development: acme.localhost
	if len(parts) == 2 && parts[1] == "localhost" {
		return parts[0]
	}

	// Production: acme.cronosplatform.com (3+ parts)
	if len(parts) >= 3 {
		return parts[0]
	}

	return ""
}

// GetTenant retrieves tenant from context
func GetTenant(ctx context.Context) *cronos.Tenant {
	tenant, _ := ctx.Value(TenantKey).(*cronos.Tenant)
	return tenant
}

// MustGetTenant retrieves tenant or panics (use in handlers where tenant is guaranteed)
func MustGetTenant(ctx context.Context) *cronos.Tenant {
	tenant := GetTenant(ctx)
	if tenant == nil {
		panic("tenant not found in context - middleware misconfigured")
	}
	return tenant
}
