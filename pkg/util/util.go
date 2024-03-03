package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"awesomeProject/pkg/data"
)

// Calculate the total amount of the order,
// including the amount and quantity of any replaced products.
func CalculateTotal(products []data.OrderProduct) string {
	total := 0.0
	for _, product := range products {
		price, _ := strconv.ParseFloat(product.Price, 64)
		total += price * float64(product.Quantity)

		// Check if the product has a replaced_with field and it's not null
		if product.ReplacedWith != nil {
			// Convert the replaced_with struct to a JSON byte slice
			replacedWithBytes, err := json.Marshal(product.ReplacedWith)
			if err != nil {
				continue
			}

			// Unmarshal the JSON byte slice into a map
			var replacedWith map[string]interface{}
			if err := json.Unmarshal(replacedWithBytes, &replacedWith); err == nil {
				if amount, ok := replacedWith["amount"].(float64); ok {
					if quantity, ok := replacedWith["quantity"].(float64); ok {
						total += amount * quantity
					}
				}
			}
		}
	}
	return fmt.Sprintf("%.2f", total)
}

// Check if order has duplicate products
func HasDuplicates(arr []int) bool {
	seen := make(map[int]bool)
	for _, value := range arr {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = true
	}
	return false
}

// Load an order by ID.
func LoadOrder(orderID string) (data.Order, bool) {
	value, ok := data.Orders.Load(orderID)
	if !ok {
		return data.Order{}, false
	}
	return value.(data.Order), true
}

// Decode JSON request body.
func DecodeJSONBody(c fiber.Ctx, v interface{}) error {
	return json.NewDecoder(bytes.NewReader(c.Body())).Decode(v)
}
