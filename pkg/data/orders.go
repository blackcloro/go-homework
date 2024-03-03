package data

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/google/uuid"
)

type OrderProduct struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Price        string        `json:"price"`
	ProductID    int           `json:"product_id"`
	Quantity     int           `json:"quantity"`
	ReplacedWith *OrderProduct `json:"replaced_with"`
}

type Amount struct {
	Discount string `json:"discount"`
	Paid     string `json:"paid"`
	Returns  string `json:"returns"`
	Total    string `json:"total"`
}

type Order struct {
	Amount   Amount         `json:"amount"`
	ID       string         `json:"id"`
	Products []OrderProduct `json:"products"`
	Status   string         `json:"status"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func NewOrder(orderID string) Order {
	return Order{
		Amount: Amount{
			Discount: "0.00",
			Paid:     "0.00",
			Returns:  "0.00",
			Total:    "0.00",
		},
		ID: orderID,

		Products: []OrderProduct{},
		Status:   "NEW",
	}
}

type UpdateProductQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type UpdateProductRequest struct {
	ReplacedWith struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	} `json:"replaced_with"`
}

func UpdateProduct(order *Order, productID string, replacementProductID int, quantity int) (string, bool) {
	for i, product := range order.Products {
		if product.ID == productID {
			for _, globalProduct := range Products {
				if globalProduct.ID == replacementProductID {
					order.Products[i].ReplacedWith = &OrderProduct{
						ID:           uuid.New().String(),
						ProductID:    globalProduct.ID,
						Name:         globalProduct.Name,
						Price:        globalProduct.Price,
						Quantity:     quantity,
						ReplacedWith: nil,
					}
					x := []OrderProduct{*order.Products[i].ReplacedWith}
					return calculateTotal(x), true
				}
			}
			break
		}
	}
	return "", false
}

func UpdateOrderAmount(order *Order, oldTotal string, newTotal string) {
	f64Old, _ := strconv.ParseFloat(oldTotal, 64)
	f64New, _ := strconv.ParseFloat(newTotal, 64)
	diff := f64Old - f64New

	if diff > 0 {
		order.Amount.Returns = strconv.FormatFloat(math.Abs(diff), 'f', 2, 64)
		order.Amount.Total = strconv.FormatFloat(f64New, 'f', 2, 64)

	} else if diff < 0 {
		order.Amount.Discount = strconv.FormatFloat(math.Abs(diff), 'f', 2, 64)
	}
}

// CalculateTotal Helper function to calculate the total amount of the order,
// including the amount and quantity of any replaced products.
func calculateTotal(products []OrderProduct) string {
	total := 0.0
	for _, product := range products {
		price, _ := strconv.ParseFloat(product.Price, 64)
		total += price * float64(product.Quantity)

		// Check if the product has a replaced_with field and it's not null
		if product.ReplacedWith != nil {
			// Convert the replaced_with struct to a JSON byte slice
			replacedWithBytes, err := json.Marshal(product.ReplacedWith)
			if err != nil {
				// Handle error
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
