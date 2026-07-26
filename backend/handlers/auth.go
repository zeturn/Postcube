package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/WorkPlace/Postcube/backend/database"
	"github.com/WorkPlace/Postcube/backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func getBasaltBaseURL() string { return getEnv("BASALT_BASE_URL", "http://localhost:8101") }
// 浏览器侧可见的 BasaltPass 地址（容器内后端访问宿主机请用 BASALT_BASE_URL=http://host.docker.internal:8101）
func getBrowserBasaltBaseURL() string {
	return getEnv("BASALT_PUBLIC_BASE_URL", "http://localhost:8101")
}
func getClientID() string      { return getEnv("BASALT_CLIENT_ID", "") }
func getClientSecret() string  { return getEnv("BASALT_CLIENT_SECRET", "") }
func getRedirectURI() string {
	return getEnv("BASALT_REDIRECT_URI", "http://localhost:8116/api/auth/callback")
}
func getFrontendURL() string       { return getEnv("FRONTEND_URL", "http://localhost:5116") }
func getSessionCookieName() string { return getEnv("SESSION_COOKIE_NAME", "postcube_session") }
func getJwtSecret() []byte         { return []byte(getEnv("JWT_SECRET", "change-this-postcube-jwt-secret")) }
func getCookieDomain() string      { return getEnv("COOKIE_DOMAIN", "") }
func getCookieSecure() bool        { return strings.EqualFold(getEnv("COOKIE_SECURE", "false"), "true") }

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func generateRandomString(length int) string {
	buf := make([]byte, length)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func generatePKCE() (verifier string, challenge string) {
	verifier = generateRandomString(32)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func setOAuthCookies(c *fiber.Ctx, state, verifier, nonce string) {
	expires := time.Now().Add(10 * time.Minute)
	secure := getCookieSecure()
	domain := getCookieDomain()

	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_state",
		Value:    state,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_verifier",
		Value:    verifier,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_nonce",
		Value:    nonce,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
}

func clearOAuthCookies(c *fiber.Ctx) {
	secure := getCookieSecure()
	domain := getCookieDomain()
	expires := time.Now().Add(-time.Hour)

	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_state",
		Value:    "",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_verifier",
		Value:    "",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "postcube_oauth_nonce",
		Value:    "",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
}

func redirectWithAuthError(c *fiber.Ctx, code string) error {
	return c.Redirect(fmt.Sprintf("%s/login?error=%s", getFrontendURL(), url.QueryEscape(code)), fiber.StatusFound)
}

func Login(c *fiber.Ctx) error {
	if strings.TrimSpace(getClientID()) == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "BASALT_CLIENT_ID is not configured"})
	}

	state := generateRandomString(16)
	nonce := generateRandomString(16)
	verifier, challenge := generatePKCE()
	setOAuthCookies(c, state, verifier, nonce)

	params := url.Values{}
	params.Set("client_id", getClientID())
	params.Set("redirect_uri", getRedirectURI())
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("nonce", nonce)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("scope", "openid profile email")

	authURL := fmt.Sprintf("%s/api/v1/oauth/authorize?%s", getBrowserBasaltBaseURL(), params.Encode())
	return c.Redirect(authURL, fiber.StatusFound)
}

func Callback(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	storedState := c.Cookies("postcube_oauth_state")
	verifier := c.Cookies("postcube_oauth_verifier")
	expectedNonce := c.Cookies("postcube_oauth_nonce")

	if code == "" || state == "" || storedState == "" || verifier == "" || storedState != state {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, "invalid_state")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", getClientID())
	if secret := strings.TrimSpace(getClientSecret()); secret != "" {
		form.Set("client_secret", secret)
	}
	form.Set("redirect_uri", getRedirectURI())
	form.Set("code", code)
	form.Set("code_verifier", verifier)

	tokenReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/oauth/token", getBasaltBaseURL()), strings.NewReader(form.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, "token_exchange_failed")
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, fmt.Sprintf("token_exchange_failed_%d", tokenResp.StatusCode))
	}

	var tokens struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil || strings.TrimSpace(tokens.AccessToken) == "" {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, "token_exchange_failed")
	}
	if strings.TrimSpace(tokens.IDToken) != "" {
		if err := validateIDTokenNonce(tokens.IDToken, expectedNonce); err != nil {
			clearOAuthCookies(c)
			return redirectWithAuthError(c, "invalid_nonce")
		}
	}

	uiReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/oauth/userinfo", getBasaltBaseURL()), nil)
	uiReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	uiResp, err := client.Do(uiReq)
	if err != nil {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, "userinfo_failed")
	}
	defer uiResp.Body.Close()

	if uiResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(uiResp.Body)
		_ = body
		clearOAuthCookies(c)
		return redirectWithAuthError(c, fmt.Sprintf("userinfo_failed_%d", uiResp.StatusCode))
	}

	var userInfo struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := json.NewDecoder(uiResp.Body).Decode(&userInfo); err != nil {
		clearOAuthCookies(c)
		return redirectWithAuthError(c, "userinfo_failed")
	}

	displayName := strings.TrimSpace(userInfo.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(userInfo.PreferredUsername)
	}
	if displayName == "" {
		emailLeft := strings.Split(strings.TrimSpace(userInfo.Email), "@")
		if len(emailLeft) > 0 && emailLeft[0] != "" {
			displayName = emailLeft[0]
		}
	}
	if displayName == "" {
		displayName = "Postcube User"
	}

	email := strings.TrimSpace(userInfo.Email)
	if email == "" {
		email = fmt.Sprintf("%s@unknown.local", strings.TrimSpace(userInfo.Sub))
	}

	var user models.User
	err = database.DB.Where("basalt_id = ?", strings.TrimSpace(userInfo.Sub)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slug, slugErr := generateUniqueSlug(displayName, 0)
			if slugErr != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to allocate user slug"})
			}
			user = models.User{
				BasaltID: strings.TrimSpace(userInfo.Sub),
				Email:    email,
				Name:     displayName,
				Slug:     slug,
				BoxTitle: fmt.Sprintf("%s's Question Box", displayName),
			}
			if err := database.DB.Create(&user).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user"})
		}
	} else {
		user.Name = displayName
		user.Email = email
		if strings.TrimSpace(user.Slug) == "" {
			slug, slugErr := generateUniqueSlug(displayName, user.ID)
			if slugErr != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user slug"})
			}
			user.Slug = slug
		}
		if strings.TrimSpace(user.BoxTitle) == "" {
			user.BoxTitle = fmt.Sprintf("%s's Question Box", displayName)
		}
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
		}
	}

	claims := jwt.MapClaims{
		"uid": user.ID,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJwtSecret())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	secure := getCookieSecure()
	domain := getCookieDomain()
	c.Cookie(&fiber.Cookie{
		Name:     getSessionCookieName(),
		Value:    signed,
		Expires:  time.Now().Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})

	clearOAuthCookies(c)
	return c.Redirect(getFrontendURL()+"/inbox", fiber.StatusFound)
}

func validateIDTokenNonce(idToken, expectedNonce string) error {
	if expectedNonce == "" {
		return errors.New("missing expected nonce")
	}
	claims := jwt.MapClaims{}
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(idToken, claims); err != nil {
		return err
	}
	if nonce, _ := claims["nonce"].(string); nonce != expectedNonce {
		return errors.New("nonce mismatch")
	}
	return nil
}

func Logout(c *fiber.Ctx) error {
	secure := getCookieSecure()
	domain := getCookieDomain()
	c.Cookie(&fiber.Cookie{
		Name:     getSessionCookieName(),
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   domain,
	})
	return c.JSON(fiber.Map{"message": "logged out"})
}

func AuthMiddleware(c *fiber.Ctx) error {
	raw := c.Cookies(getSessionCookieName())
	if strings.TrimSpace(raw) == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return getJwtSecret(), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID, ok := claimToUint(claims["uid"])
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	c.Locals("user_id", userID)
	return c.Next()
}

func Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

func claimToUint(v interface{}) (uint, bool) {
	switch n := v.(type) {
	case float64:
		return uint(n), true
	case int:
		return uint(n), true
	case int64:
		return uint(n), true
	case uint:
		return n, true
	case string:
		parsed, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(parsed), true
	default:
		return 0, false
	}
}

func generateUniqueSlug(seed string, currentUserID uint) (string, error) {
	base := slugify(seed)
	if base == "" {
		base = "postcube-user"
	}

	candidate := base
	for i := 0; i < 2000; i++ {
		var count int64
		q := database.DB.Model(&models.User{}).Where("slug = ?", candidate)
		if currentUserID > 0 {
			q = q.Where("id <> ?", currentUserID)
		}
		if err := q.Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}

	return "", errors.New("failed to generate unique slug")
}

func slugify(input string) string {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range trimmed {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if len(out) > 64 {
		out = strings.Trim(out[:64], "-")
	}
	return out
}
