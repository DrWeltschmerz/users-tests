package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DrWeltschmerz/jwt-auth/pkg/authjwt"
	ginadapter "github.com/DrWeltschmerz/users-adapter-gin/ginadapter"
	gormAdapter "github.com/DrWeltschmerz/users-adapter-gorm/gorm"
	"github.com/DrWeltschmerz/users-core"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testDBInit bool
)

func setupGinIntegrationServer(t *testing.T) (*gin.Engine, func()) {
	if !testDBInit {
		var err error
		testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)
		err = testDB.AutoMigrate(&gormAdapter.GormUser{}, &gormAdapter.GormRole{})
		require.NoError(t, err)
		testDBInit = true
	}

	userRepo := gormAdapter.NewGormUserRepository(testDB)
	roleRepo := gormAdapter.NewGormRoleRepository(testDB)
	hasher := authjwt.NewBcryptHasher()
	// Set JWT_SECRET environment variable for the tokenizer
	t.Setenv("JWT_SECRET", "test-secret-very-long-and-secure")
	tokenizer := authjwt.NewJWTTokenizer()

	svc := users.NewService(userRepo, roleRepo, hasher, tokenizer)

	r := gin.Default()
	ginadapter.RegisterRoutes(r, svc, tokenizer)

	cleanup := func() {
		// Do not close testDB to allow reuse across tests
	}

	return r, cleanup
}

func TestGinIntegration_Register_Login_Profile(t *testing.T) {

	r, cleanup := setupGinIntegrationServer(t)
	defer cleanup()

	// Register
	registerBody := map[string]string{"email": "gin@ex.com", "username": "ginuser", "password": "pw123"}
	b, _ := json.Marshal(registerBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var regUser users.User
	json.Unmarshal(w.Body.Bytes(), &regUser)

	// Login
	loginBody := map[string]string{"email": "gin@ex.com", "password": "pw123"}
	b, _ = json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var loginResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]
	require.NotEmpty(t, token)

	// Get profile (authenticated)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var user users.User
	json.Unmarshal(w.Body.Bytes(), &user)
	require.Equal(t, "gin@ex.com", user.Email)

	// Update profile
	updateBody := map[string]string{"username": "ginuser2"}
	b, _ = json.Marshal(updateBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/user/profile", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var updatedUser users.User
	json.Unmarshal(w.Body.Bytes(), &updatedUser)
	require.Equal(t, "ginuser2", updatedUser.Username)

	// Change password
	changePwBody := map[string]string{"old_password": "pw123", "new_password": "pw456"}
	b, _ = json.Marshal(changePwBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/user/change-password", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Login with new password
	loginBody = map[string]string{"email": "gin@ex.com", "password": "pw456"}
	b, _ = json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token = loginResp["token"]
	require.NotEmpty(t, token)

	// List users (should fail for normal user)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// List roles (should fail for normal user)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/roles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// Try to assign role (should fail for normal user)
	assignBody := map[string]string{"role_id": "1"}
	b, _ = json.Marshal(assignBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/users/1/assign-role", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// Try to reset password (should fail for normal user)
	resetBody := map[string]string{"new_password": "pw789"}
	b, _ = json.Marshal(resetBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/users/1/reset-password", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// Try to delete user (should fail for normal user)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// Negative: Register with duplicate email
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/register", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusCreated, w.Code)

	// Negative: Login with wrong password
	wrongLogin := map[string]string{"email": "gin@ex.com", "password": "wrongpass"}
	b2, _ := json.Marshal(wrongLogin)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewReader(b2))
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Change password with wrong old password
	badPwBody := map[string]string{"old_password": "badold", "new_password": "pw999"}
	b2, _ = json.Marshal(badPwBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/user/change-password", bytes.NewReader(b2))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Access profile without token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/user/profile", nil)
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Try to update another user (should fail, endpoint not exposed, but test for 404/403)
	updateBody = map[string]string{"username": "hacker"}
	b2, _ = json.Marshal(updateBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/users/2", bytes.NewReader(b2))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Register with missing fields
	missingBody := map[string]string{"email": "", "username": "", "password": ""}
	b3, _ := json.Marshal(missingBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/register", bytes.NewReader(b3))
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusCreated, w.Code)

	// Negative: Login with non-existent user
	nonExistLogin := map[string]string{"email": "nope@ex.com", "password": "pw123"}
	b3, _ = json.Marshal(nonExistLogin)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewReader(b3))
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Access admin endpoint with invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)

	// Negative: Access protected endpoint with expired/invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/user/profile", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	r.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusOK, w.Code)
}
