package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"awesomeProject/pkg/api"
	"awesomeProject/pkg/data"

	"github.com/gofiber/fiber/v3"
)

const (
	apiProductsPath = "/api/products"
	apiOrdersPath   = "/api/orders"
	referenceAPIURL = "https://homework.solutional.ee"
)

// Test GET /api/products - list of all available products.
func TestGetProductsCompare(t *testing.T) {
	t.Parallel()
	app := setupApp()

	resp := performRequestAndCheckStatus(t, app, fiber.MethodGet, apiProductsPath, nil, http.StatusOK)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	refBody, err := getReferenceBody(referenceAPIURL + apiProductsPath)
	if err != nil {
		t.Fatalf("failed to get reference body: %v", err)
	}

	if !reflect.DeepEqual(body, refBody) {
		t.Errorf("handler returned different body than reference API: got %v want %v", body, refBody)
	}
}

// Test POST /api/orders - create a new order &
// Test GET /api/orders/:order_id - get order details.
func TestCreatingOrder(t *testing.T) {
	t.Parallel()
	app := setupApp()

	newOrder, refOrder := createOrders(t, app)

	checkOrder := getOrder(t, app, newOrder.ID)
	checkRefOrder := getReferenceOrder(t, refOrder.ID)

	// Third argument is for orders with products
	if !ordersEqual(checkOrder, checkRefOrder, false) {
		t.Errorf("handler returned different body than reference API: got %v want %v", checkOrder, checkRefOrder)
	}
}

// Test PATCH /api/orders/:order_id - update an order
func TestUpdateOrderStatus(t *testing.T) {
	t.Parallel()
	app := setupApp()

	newOrder, refOrder := createOrders(t, app)

	// Update both orders to "PAID"
	updateOrderStatus(t, app, newOrder.ID, "PAID", false)
	updateOrderStatus(t, app, refOrder.ID, "PAID", true)

	updatedOrder := getOrder(t, app, newOrder.ID)
	updatedRefOrder := getReferenceOrder(t, refOrder.ID)

	// Adjusted expectation to "PAID"
	if updatedOrder.Status != "PAID" || updatedRefOrder.Status != "PAID" {
		t.Errorf("handler returned different body than reference API: got %v want %v",
			updatedOrder, updatedRefOrder)
	}
}

// Test POST /api/orders/:order_id/products - add products to the order
func TestAddProductToOrder(t *testing.T) {
	t.Parallel()
	app := setupApp()

	// Step 1: Create an order
	newOrder, refOrder := createOrders(t, app)

	addProduct(t, app, newOrder.ID, "123", false)
	addProduct(t, app, refOrder.ID, "123", true)

	updatedOrder := getOrder(t, app, newOrder.ID)
	updatedRefOrder := getReferenceOrder(t, refOrder.ID)

	if !ordersEqual(updatedOrder, updatedRefOrder, true) {
		t.Errorf("handler returned different body than reference API: got %v want %v", updatedOrder, updatedRefOrder)
	}
}

// Test POST /api/orders/:order_id/products - add products to the order
// Discount scenario
func TestReplaceProductToOrderDiscount(t *testing.T) {
	t.Parallel()
	app := setupApp()

	// Step 1: Create an order
	newOrder, refOrder := createOrders(t, app)

	addProduct(t, app, newOrder.ID, "123", false)
	addProduct(t, app, refOrder.ID, "123", true)

	updatedOrderID := getOrder(t, app, newOrder.ID).Products[0].ID
	updatedRefOrderID := getReferenceOrder(t, refOrder.ID).Products[0].ID

	// Update both orders to "PAID"
	updateOrderStatus(t, app, newOrder.ID, "PAID", false)
	updateOrderStatus(t, app, refOrder.ID, "PAID", true)

	replaceProduct(t, app, newOrder.ID, updatedOrderID, "123", false)
	replaceProduct(t, app, refOrder.ID, updatedRefOrderID, "123", true)

	updatedOrderAmount := getOrder(t, app, newOrder.ID).Amount
	updatedRefOrderAmount := getReferenceOrder(t, refOrder.ID).Amount
	if !reflect.DeepEqual(updatedOrderAmount, updatedRefOrderAmount) {
		t.Errorf("handler returned different body than reference API: got %v want %v", updatedOrderAmount, updatedRefOrderAmount)
	}
}

// Return scenario
func TestReplaceProductToOrderReturn(t *testing.T) {
	t.Parallel()
	app := setupApp()

	// Step 1: Create an order
	newOrder, refOrder := createOrders(t, app)

	addProduct(t, app, newOrder.ID, "999", false)
	addProduct(t, app, refOrder.ID, "999", true)

	updatedOrderID := getOrder(t, app, newOrder.ID).Products[0].ID
	updatedRefOrderID := getReferenceOrder(t, refOrder.ID).Products[0].ID

	// Update both orders to "PAID"
	updateOrderStatus(t, app, newOrder.ID, "PAID", false)
	updateOrderStatus(t, app, refOrder.ID, "PAID", true)

	replaceProduct(t, app, newOrder.ID, updatedOrderID, "123", false)
	replaceProduct(t, app, refOrder.ID, updatedRefOrderID, "123", true)

	updatedOrderAmount := getOrder(t, app, newOrder.ID).Amount
	updatedRefOrderAmount := getReferenceOrder(t, refOrder.ID).Amount
	if !reflect.DeepEqual(updatedOrderAmount, updatedRefOrderAmount) {
		t.Errorf("handler returned different body than reference API: got %v want %v", updatedOrderAmount, updatedRefOrderAmount)
	}
}

// Setup testing server for API.
func setupApp() *fiber.App {
	app := fiber.New()
	registerHandlers(app)
	return app
}

func registerHandlers(app *fiber.App) {
	app.Get(apiProductsPath, api.GetProducts)
	app.Post(apiOrdersPath, api.CreateOrder)
	app.Get(apiOrdersPath+"/:order_id", api.GetOrder)
	app.Patch(apiOrdersPath+"/:order_id", api.UpdateOrderStatus)
	app.Post("/api/orders/:order_id/products", api.AddProductsToOrder)
	app.Patch("/api/orders/:order_id/products/:product_id", api.ProductPatchHandler)
}

func performRequestAndCheckStatus(t *testing.T, app *fiber.App, method, path string, body io.Reader, expectedStatus int) *http.Response {
	req := httptest.NewRequest(method, path, body)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}
	checkStatusCode(t, resp, expectedStatus)
	return resp
}

func checkStatusCode(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if status := resp.StatusCode; status != want {
		t.Errorf("handler returned wrong status code: got %v want %v", status, want)
	}
}

func createOrders(t *testing.T, app *fiber.App) (data.Order, data.Order) {
	t.Helper()
	resp := performRequestAndCheckStatus(t, app, fiber.MethodPost, apiOrdersPath, nil, http.StatusCreated)
	defer resp.Body.Close()

	var newOrder data.Order
	unmarshalResponseBody(t, resp, &newOrder)

	refResp, err := http.Post(referenceAPIURL+apiOrdersPath, "application/json", nil)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	defer refResp.Body.Close()

	var refOrder data.Order
	unmarshalResponseBody(t, refResp, &refOrder)

	return newOrder, refOrder
}

func getOrder(t *testing.T, app *fiber.App, orderID string) data.Order {
	t.Helper()
	resp := performRequestAndCheckStatus(t, app, fiber.MethodGet, apiOrdersPath+"/"+orderID, nil, http.StatusOK)
	defer resp.Body.Close()

	var order data.Order
	unmarshalResponseBody(t, resp, &order)
	return order
}

func getReferenceOrder(t *testing.T, orderID string) data.Order {
	t.Helper()
	refResp, err := http.Get(referenceAPIURL + apiOrdersPath + "/" + orderID)
	if err != nil {
		t.Fatalf("failed to create reference request: %v", err)
	}
	defer refResp.Body.Close()

	var refOrder data.Order
	unmarshalResponseBody(t, refResp, &refOrder)
	return refOrder
}

func unmarshalResponseBody(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
}

func ordersEqual(a, b data.Order, products bool) bool {
	// IDs are not part of the equality check, as it's generated and expected to differ
	a.ID = ""
	b.ID = ""
	if products {
		for i := range a.Products {
			a.Products[i].ID = ""
			b.Products[i].ID = ""
		}
	}
	return reflect.DeepEqual(a, b)
}

func getReferenceBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Accepts a flag to determine whether to update the local or reference API order.
func updateOrderStatus(t *testing.T, app *fiber.App, orderID, status string, isReference bool) {
	t.Helper()
	updateRequest := data.UpdateOrderStatusRequest{Status: status}
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var url string
	if isReference {
		url = fmt.Sprintf("%s%s/%s", referenceAPIURL, apiOrdersPath, orderID)
	} else {
		url = fmt.Sprintf("/api/orders/%s", orderID)
	}

	resp, err := makeRequest(t, app, http.MethodPatch, url, bytes.NewBuffer(requestBody), isReference)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}
	defer resp.Body.Close()

	checkStatusCode(t, resp, http.StatusOK)
}

// Accepts a flag to determine whether to update the local or reference API order.
func addProduct(t *testing.T, app *fiber.App, orderID, productID string, isReference bool) {
	t.Helper()
	updateRequest := fmt.Sprintf("[%s]", productID)
	requestBody := bytes.NewBufferString(updateRequest)

	var url string
	if isReference {
		url = fmt.Sprintf("%s%s/%s/products", referenceAPIURL, apiOrdersPath, orderID)
	} else {
		url = fmt.Sprintf("/api/orders/%s/products", orderID)
	}

	resp, err := makeRequest(t, app, http.MethodPost, url, requestBody, isReference)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}
	defer resp.Body.Close()

	checkStatusCode(t, resp, http.StatusCreated)
}

// Accepts a flag to determine whether to update the local or reference API order.
func replaceProduct(t *testing.T, app *fiber.App, orderID, productID, replacementProductID string, isReference bool) {
	t.Helper()
	updateRequest := fmt.Sprintf("{\"replaced_with\": {\"product_id\": %s, \"quantity\": 6}}", replacementProductID)
	requestBody := bytes.NewBufferString(updateRequest)

	var url string
	if isReference {
		url = fmt.Sprintf("%s%s/%s/products/%s", referenceAPIURL, apiOrdersPath, orderID, productID)
	} else {
		url = fmt.Sprintf("/api/orders/%s/products/%s", orderID, productID)
	}

	resp, err := makeRequest(t, app, http.MethodPatch, url, requestBody, isReference)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}
	defer resp.Body.Close()

	checkStatusCode(t, resp, http.StatusOK)
}

// Logic for making HTTP requests, which is shared between local and reference API updates.
func makeRequest(t *testing.T, app *fiber.App, method, path string, body io.Reader, isReference bool) (*http.Response, error) {
	t.Helper()
	if isReference {
		client := &http.Client{}
		req, err := http.NewRequest(method, path, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return client.Do(req)
	}
	return app.Test(httptest.NewRequest(method, path, body))
}
