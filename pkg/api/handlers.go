package api

import (
	"encoding/json"

	"awesomeProject/pkg/util"

	"awesomeProject/pkg/data"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const PAID = "PAID"

// GetProducts retrieves all products.
func GetProducts(c fiber.Ctx) error {
	return c.JSON(data.Products)
}

func CreateOrder(c fiber.Ctx) error {
	oID, _ := uuid.NewV7()
	orderID := oID.String()

	order := data.NewOrder(orderID)
	data.Orders.Store(orderID, order)

	return c.Status(fiber.StatusCreated).JSON(order)
}

func GetOrder(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}
	return c.JSON(order)
}

// UpdateOrderStatus updates the status of an existing order.
func UpdateOrderStatus(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	var request data.UpdateOrderStatusRequest
	if err := util.DecodeJSONBody(c, &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	if request.Status != "NEW" && request.Status != PAID {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid order status")
	}

	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}

	if order.Status == request.Status || order.Status == PAID {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid order status")
	}

	order.Status = request.Status
	if request.Status == PAID {
		order.Amount.Paid = order.Amount.Total
	}
	data.Orders.Store(orderID, order)

	return c.Status(fiber.StatusOK).JSON("OK")
}

func AddProductsToOrder(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	var productIDs []int
	if err := util.DecodeJSONBody(c, &productIDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	if util.HasDuplicates(productIDs) {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}

	for _, id := range productIDs {
		found := false
		for i, product := range order.Products {
			if product.ProductID == id {
				// Product already exists, increment its quantity by 1
				product.Quantity++
				order.Products[i] = product
				found = true
				break
			}
		}

		if !found {
			// If product not found, find it in the global products list and add to order
			for _, globalProduct := range data.Products {
				if globalProduct.ID == id {
					id := uuid.New().String()
					order.Products = append(order.Products, data.OrderProduct{
						ID:           id,
						ProductID:    globalProduct.ID,
						Name:         globalProduct.Name,
						Price:        globalProduct.Price,
						Quantity:     1,
						ReplacedWith: nil,
					})
					break
				}
			}
		}
	}

	// Update the Total field in the Amount struct
	order.Amount.Total = util.CalculateTotal(order.Products)

	data.Orders.Store(orderID, order)
	return c.Status(fiber.StatusCreated).JSON("OK")
}

// GetOrderProducts retrieves the products of an order.
func GetOrderProducts(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}

	if len(order.Products) == 0 {
		return c.JSON([]data.Product{})
	}

	return c.JSON(order.Products)
}

// UpdateProductQuantity updates the quantity of a product in an order and recalculates the total amount.
func UpdateProductQuantity(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	productID := c.Params("product_id")
	var request data.UpdateProductQuantityRequest
	if err := util.DecodeJSONBody(c, &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}

	// Validate the status of order
	if order.Status == PAID {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	for i, product := range order.Products {
		if product.ID == productID {
			// Update the product's quantity
			order.Products[i].Quantity = request.Quantity
			// Recalculate the total amount of the order
			order.Amount.Total = util.CalculateTotal(order.Products)
			data.Orders.Store(orderID, order)
			return c.Status(fiber.StatusOK).JSON("OK")
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not Found"})
}

func AddReplacementProduct(c fiber.Ctx) error {
	orderID := c.Params("order_id")
	productID := c.Params("product_id")

	type UpdateProductRequest struct {
		ReplacedWith struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		} `json:"replaced_with"`
	}
	var request UpdateProductRequest

	if err := util.DecodeJSONBody(c, &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	if request.ReplacedWith.Quantity < 1 {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	order, ok := util.LoadOrder(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}

	// Ensure all calculations and updates to the order's financial data are done within this critical section.
	oldTotal := order.Amount.Total
	newTotal, found := data.UpdateProduct(&order, productID, request.ReplacedWith.ProductID, request.ReplacedWith.Quantity)

	if !found {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}
	data.UpdateOrderAmount(&order, oldTotal, newTotal)
	data.Orders.Store(orderID, order)

	return c.Status(fiber.StatusOK).JSON("OK")
}

func ProductPatchHandler(c fiber.Ctx) error {
	var body map[string]interface{}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid parameters")
	}

	switch {
	case body["replaced_with"] != nil:
		return AddReplacementProduct(c)
	case body["quantity"] != nil:
		return UpdateProductQuantity(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON("Invalid action")
	}
}
