package data

import "sync"

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

var (
	Products = []Product{
		{ID: 123, Name: "Ketchup", Price: "0.45"},
		{ID: 456, Name: "Beer", Price: "2.33"},
		{ID: 879, Name: "Õllesnäkk", Price: "0.42"},
		{ID: 999, Name: "75\" OLED TV", Price: "1333.37"},
	}
	Orders = sync.Map{}
)
