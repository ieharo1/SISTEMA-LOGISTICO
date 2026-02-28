package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"dispatchpro/internal/models"
	"dispatchpro/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrProductNotFound = errors.New("product not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrDriverNotFound  = errors.New("driver not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repositories.NewUserRepository(),
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hash),
		Name:     req.Name,
		Role:     role,
		Active:   true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	user.Password = ""
	return user, nil
}

func (s *AuthService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

var stockMutex sync.Mutex

type ProductService struct {
	productRepo    *repositories.ProductRepository
	inventoryLogRepo *repositories.InventoryLogRepository
}

func NewProductService() *ProductService {
	return &ProductService{
		productRepo: repositories.NewProductRepository(),
		inventoryLogRepo: repositories.NewInventoryLogRepository(),
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product, userID primitive.ObjectID) (*models.Product, error) {
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	s.recordInventoryLog(ctx, product.ID, "created", product.Stock, product.Stock, "Initial stock", userID)
	return product, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	return s.productRepo.FindAll(ctx)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.productRepo.Update(ctx, product)
}

func (s *ProductService) AdjustStock(ctx context.Context, productID primitive.ObjectID, quantity int, reason string, userID primitive.ObjectID) error {
	stockMutex.Lock()
	defer stockMutex.Unlock()

	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	previousStock := product.Stock
	newStock := previousStock + quantity

	if newStock < 0 {
		return ErrInsufficientStock
	}

	if err := s.productRepo.UpdateStock(ctx, productID, newStock); err != nil {
		return err
	}

	action := "increase"
	if quantity < 0 {
		action = "decrease"
	}
	s.recordInventoryLog(ctx, productID, action, previousStock, newStock, reason, userID)

	return nil
}

func (s *ProductService) GetLowStockProducts(ctx context.Context) ([]models.Product, error) {
	return s.productRepo.GetLowStock(ctx)
}

func (s *ProductService) recordInventoryLog(ctx context.Context, productID primitive.ObjectID, action string, previousStock, newStock int, reason string, userID primitive.ObjectID) {
	log := &models.InventoryLog{
		ProductID:    productID,
		Action:       action,
		Quantity:     newStock - previousStock,
		PreviousStock: previousStock,
		NewStock:     newStock,
		Reason:       reason,
		UserID:       userID,
	}
	s.inventoryLogRepo.Create(ctx, log)
}

func (s *ProductService) GetInventoryHistory(ctx context.Context, productID primitive.ObjectID) ([]models.InventoryLog, error) {
	return s.inventoryLogRepo.FindByProductID(ctx, productID)
}

type OrderService struct {
	orderRepo    *repositories.OrderRepository
	productRepo  *repositories.ProductRepository
	driverRepo   *repositories.DriverRepository
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo:   repositories.NewOrderRepository(),
		productRepo: repositories.NewProductRepository(),
		driverRepo:  repositories.NewDriverRepository(),
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order, userID primitive.ObjectID) error {
	for i := range order.Items {
		product, err := s.productRepo.FindByID(ctx, order.Items[i].ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return fmt.Errorf("product not found: %s", order.Items[i].ProductID)
		}

		order.Items[i].Name = product.Name
		order.Items[i].UnitPrice = product.Price
		order.Items[i].Subtotal = product.Price * float64(order.Items[i].Quantity)

		stockMutex.Lock()
		if product.Stock < order.Items[i].Quantity {
			stockMutex.Unlock()
			return fmt.Errorf("insufficient stock for product: %s", product.Name)
		}
		newStock := product.Stock - order.Items[i].Quantity
		s.productRepo.UpdateStock(ctx, product.ID, newStock)
		stockMutex.Unlock()
	}

	order.Total = 0
	for _, item := range order.Items {
		order.Total += item.Subtotal
	}

	order.Status = models.OrderPending
	return s.orderRepo.Create(ctx, order)
}

func (s *OrderService) GetOrder(ctx context.Context, id primitive.ObjectID) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) GetOrders(ctx context.Context, status string, page, limit int) ([]models.Order, int64, error) {
	return s.orderRepo.Find(ctx, status, page, limit)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID primitive.ObjectID, status models.OrderStatus) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	order.Status = status
	if status == models.OrderDelivered {
		now := time.Now()
		order.DeliveredAt = &now
	}

	return s.orderRepo.Update(ctx, order)
}

func (s *OrderService) AssignDriver(ctx context.Context, orderID, driverID primitive.ObjectID) error {
	driver, err := s.driverRepo.FindByID(ctx, driverID)
	if err != nil {
		return err
	}
	if driver == nil {
		return ErrDriverNotFound
	}

	return s.orderRepo.AssignDriver(ctx, orderID, driverID)
}

func (s *OrderService) GetStats(ctx context.Context) (*models.DashboardStats, error) {
	stats, err := s.orderRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	activeDrivers, _ := s.driverRepo.GetActiveCount(ctx)
	stats.ActiveDrivers = activeDrivers

	lowStock, _ := s.productRepo.GetLowStock(ctx)
	stats.LowStockItems = len(lowStock)

	return stats, nil
}

type DriverService struct {
	driverRepo *repositories.DriverRepository
}

func NewDriverService() *DriverService {
	return &DriverService{
		driverRepo: repositories.NewDriverRepository(),
	}
}

func (s *DriverService) CreateDriver(ctx context.Context, driver *models.Driver) (*models.Driver, error) {
	if err := s.driverRepo.Create(ctx, driver); err != nil {
		return nil, err
	}
	return driver, nil
}

func (s *DriverService) GetDriver(ctx context.Context, id primitive.ObjectID) (*models.Driver, error) {
	driver, err := s.driverRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if driver == nil {
		return nil, ErrDriverNotFound
	}
	return driver, nil
}

func (s *DriverService) GetAllDrivers(ctx context.Context) ([]models.Driver, error) {
	return s.driverRepo.FindAll(ctx)
}

func (s *DriverService) UpdateDriver(ctx context.Context, driver *models.Driver) error {
	return s.driverRepo.Update(ctx, driver)
}
