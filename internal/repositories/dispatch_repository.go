package repositories

import (
	"context"
	"fmt"
	"time"

	"dispatchpro/internal/config"
	"dispatchpro/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() *UserRepository {
	db := config.GetDatabase()
	return &UserRepository{
		collection: db.Database.Collection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *UserRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository() *ProductRepository {
	db := config.GetDatabase()
	return &ProductRepository{
		collection: db.Database.Collection("products"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	var product models.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindAll(ctx context.Context) ([]models.Product, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": product.ID}, bson.M{"$set": product})
	return err
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id primitive.ObjectID, newStock int) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"stock": newStock, "updated_at": time.Now()}})
	return err
}

func (r *ProductRepository) GetLowStock(ctx context.Context) ([]models.Product, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"active": true, "$expr": bson.M{"$lte": []interface{}{"$stock", "$min_stock"}}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "sku", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository() *OrderRepository {
	db := config.GetDatabase()
	return &OrderRepository{
		collection: db.Database.Collection("orders"),
	}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Find(ctx context.Context, status string, page, limit int) ([]models.Order, int64, error) {
	query := bson.M{}
	if status != "" {
		query["status"] = status
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	order.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": order.ID}, bson.M{"$set": order})
	return err
}

func (r *OrderRepository) AssignDriver(ctx context.Context, orderID, driverID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": orderID}, bson.M{
		"$set": bson.M{
			"driver_id":   driverID,
			"status":      models.OrderShipped,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *OrderRepository) GetStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{
		OrdersByStatus: make(map[string]int),
	}

	total, _ := r.collection.CountDocuments(ctx, bson.M{})
	stats.TotalOrders = int(total)

	pending, _ := r.collection.CountDocuments(ctx, bson.M{"status": models.OrderPending})
	stats.PendingOrders = int(pending)

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
	}

	cursor, _ := r.collection.Aggregate(ctx, pipeline)
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		cursor.Decode(&result)
		stats.OrdersByStatus[result.ID] = result.Count
	}

	revenuePipeline := []bson.M{
		{"$match": bson.M{"status": bson.M{"$in": []string{"delivered", "shipped"}}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$total"}}},
	}

	cursor, _ = r.collection.Aggregate(ctx, revenuePipeline)
	if cursor.Next(ctx) {
		var result struct {
			Total float64 `bson:"total"`
		}
		cursor.Decode(&result)
		stats.TotalRevenue = result.Total
	}

	return stats, nil
}

func (r *OrderRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "order_number", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "driver_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type DriverRepository struct {
	collection *mongo.Collection
}

func NewDriverRepository() *DriverRepository {
	db := config.GetDatabase()
	return &DriverRepository{
		collection: db.Database.Collection("drivers"),
	}
}

func (r *DriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	driver.CreatedAt = time.Now()
	driver.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, driver)
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	driver.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *DriverRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Driver, error) {
	var driver models.Driver
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&driver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &driver, nil
}

func (r *DriverRepository) FindAll(ctx context.Context) ([]models.Driver, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []models.Driver
	if err := cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}
	return drivers, nil
}

func (r *DriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	driver.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": driver.ID}, bson.M{"$set": driver})
	return err
}

func (r *DriverRepository) GetActiveCount(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"active": true})
	return int(count), err
}

func (r *DriverRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "license", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type InventoryLogRepository struct {
	collection *mongo.Collection
}

func NewInventoryLogRepository() *InventoryLogRepository {
	db := config.GetDatabase()
	return &InventoryLogRepository{
		collection: db.Database.Collection("inventory_logs"),
	}
}

func (r *InventoryLogRepository) Create(ctx context.Context, log *models.InventoryLog) error {
	log.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create inventory log: %w", err)
	}
	log.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *InventoryLogRepository) FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]models.InventoryLog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50)
	cursor, err := r.collection.Find(ctx, bson.M{"product_id": productID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []models.InventoryLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

type RouteRepository struct {
	collection *mongo.Collection
}

func NewRouteRepository() *RouteRepository {
	db := config.GetDatabase()
	return &RouteRepository{
		collection: db.Database.Collection("routes"),
	}
}

func (r *RouteRepository) Create(ctx context.Context, route *models.Route) error {
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, route)
	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}
	route.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *RouteRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Route, error) {
	var route models.Route
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&route)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

func (r *RouteRepository) Update(ctx context.Context, route *models.Route) error {
	route.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": route.ID}, bson.M{"$set": route})
	return err
}

func (r *RouteRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "driver_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
