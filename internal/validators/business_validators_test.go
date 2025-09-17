package validators

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crm_lite/internal/dto"
	"crm_lite/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCustomValidators(t *testing.T) {
	// Register both mobile and custom validators
	validator.RegisterMobileValidator()
	RegisterCustomValidators()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test endpoint for order creation
	router.POST("/orders", func(c *gin.Context) {
		var req dto.OrderCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test endpoint for wallet transaction
	router.POST("/wallet/transactions", func(c *gin.Context) {
		var req dto.WalletTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test endpoint for customer creation
	router.POST("/customers", func(c *gin.Context) {
		var req dto.CustomerCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("ValidOrderStatus", func(t *testing.T) {
		body := `{
			"customer_id": 1,
			"order_date": "2024-01-01T10:00:00Z",
			"status": "draft",
			"items": [{"product_id": 1, "quantity": 1, "unit_price": 100.0}]
		}`

		req := httptest.NewRequest("POST", "/orders", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidOrderStatus", func(t *testing.T) {
		body := `{
			"customer_id": 1,
			"order_date": "2024-01-01T10:00:00Z",
			"status": "invalid_status",
			"items": [{"product_id": 1, "quantity": 1, "unit_price": 100.0}]
		}`

		req := httptest.NewRequest("POST", "/orders", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "order_create_status")
	})

	t.Run("ValidWalletTransactionType", func(t *testing.T) {
		body := `{
			"amount": 100.0,
			"type": "recharge",
			"source": "manual",
			"remark": "test recharge"
		}`

		req := httptest.NewRequest("POST", "/wallet/transactions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidWalletTransactionType", func(t *testing.T) {
		body := `{
			"amount": 100.0,
			"type": "invalid_type",
			"source": "manual",
			"remark": "test transaction"
		}`

		req := httptest.NewRequest("POST", "/wallet/transactions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "wallet_transaction_type")
	})

	t.Run("ValidCustomerGender", func(t *testing.T) {
		body := `{
			"name": "Test User",
			"phone": "13800138000",
			"gender": "male"
		}`

		req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidCustomerGender", func(t *testing.T) {
		body := `{
			"name": "Test User",
			"phone": "13800138000",
			"gender": "invalid_gender"
		}`

		req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "customer_gender")
	})

	t.Run("EmptyOptionalFields", func(t *testing.T) {
		// Test that empty optional fields are allowed
		body := `{
			"name": "Test User",
			"phone": "13800138000",
			"gender": "",
			"level": "",
			"source": ""
		}`

		req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}