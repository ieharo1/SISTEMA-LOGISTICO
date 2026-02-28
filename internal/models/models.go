package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Name      string             `bson:"name" json:"name"`
	Role      string             `bson:"role" json:"role"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	SKU         string             `bson:"sku" json:"sku"`
	Price       float64            `bson:"price" json:"price"`
	Weight      float64            `bson:"weight" json:"weight"`
	Dimensions  string             `bson:"dimensions" json:"dimensions"`
	Stock       int                `bson:"stock" json:"stock"`
	MinStock    int                `bson:"min_stock" json:"min_stock"`
	Category    string             `bson:"category" json:"category"`
	Active      bool               `bson:"active" json:"active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderConfirmed OrderStatus = "confirmed"
	OrderProcessing OrderStatus = "processing"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	OrderNumber   string               `bson:"order_number" json:"order_number"`
	CustomerName  string               `bson:"customer_name" json:"customer_name"`
	CustomerEmail string               `bson:"customer_email" json:"customer_email"`
	CustomerPhone string               `bson:"customer_phone" json:"customer_phone"`
	ShippingAddr  string               `bson:"shipping_addr" json:"shipping_addr"`
	Items         []OrderItem          `bson:"items" json:"items"`
	Total         float64              `bson:"total" json:"total"`
	Status        OrderStatus          `bson:"status" json:"status"`
	DriverID      *primitive.ObjectID  `bson:"driver_id,omitempty" json:"driver_id,omitempty"`
	RouteID       *primitive.ObjectID  `bson:"route_id,omitempty" json:"route_id,omitempty"`
	Notes         string               `bson:"notes" json:"notes"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
	DeliveredAt   *time.Time           `bson:"delivered_at,omitempty" json:"delivered_at,omitempty"`
}

type OrderItem struct {
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id"`
	Name      string             `bson:"name" json:"name"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	UnitPrice float64            `bson:"unit_price" json:"unit_price"`
	Subtotal  float64            `bson:"subtotal" json:"subtotal"`
}

type Driver struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Phone     string             `bson:"phone" json:"phone"`
	License   string             `bson:"license" json:"license"`
	Vehicle   string             `bson:"vehicle" json:"vehicle"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Route struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	DriverID    primitive.ObjectID `bson:"driver_id" json:"driver_id"`
	OrderIDs    []primitive.ObjectID `bson:"order_ids" json:"order_ids"`
	Status      string             `bson:"status" json:"status"`
	StartTime   time.Time          `bson:"start_time" json:"start_time"`
	EndTime     *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	TotalDistance float64          `bson:"total_distance" json:"total_distance"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type InventoryLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id"`
	Action      string             `bson:"action" json:"action"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	PreviousStock int              `bson:"previous_stock" json:"previous_stock"`
	NewStock    int                `bson:"new_stock" json:"new_stock"`
	Reason      string             `bson:"reason" json:"reason"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type DashboardStats struct {
	TotalOrders    int              `json:"total_orders"`
	PendingOrders  int              `json:"pending_orders"`
	ActiveDrivers int              `json:"active_drivers"`
	LowStockItems  int              `json:"low_stock_items"`
	TotalRevenue   float64          `json:"total_revenue"`
	OrdersByStatus map[string]int   `json:"orders_by_status"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
