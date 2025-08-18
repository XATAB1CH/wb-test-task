package db

import (
	"context"
	"fmt"
	"wb-test-task/config"
	"wb-test-task/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewPostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveOrder(ctx context.Context, order models.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// Сохранение основной информации о заказе
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, 
			internal_signature, customer_id, delivery_service, 
			shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		return fmt.Errorf("insert order failed: %w", err)
	}

	// Сохранение данных о доставке
	_, err = tx.Exec(ctx, `
		INSERT INTO deliveries (order_uid, name, phone, zip, city, 
			address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery failed: %w", err)
	}

	// Сохранение данных об оплате
	_, err = tx.Exec(ctx, `
		INSERT INTO payments (order_uid, transaction, request_id, currency, 
			provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment failed: %w", err)
	}

	// Сохранение товаров
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, 
				name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NMID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("insert item failed: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	var order models.Order

	// 1. Получаем основную информацию о заказе
	err = tx.QueryRow(ctx, `
        SELECT 
            order_uid, track_number, entry, locale, 
            internal_signature, customer_id, delivery_service, 
            shardkey, sm_id, date_created, oof_shard
        FROM public.orders 
        WHERE order_uid = $1`, orderUID).
		Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
		)
	if err != nil {
		return nil, fmt.Errorf("select order failed: %w", err)
	}

	// 2. Получаем информацию о доставке
	err = tx.QueryRow(ctx, `
        SELECT 
            name, phone, zip, city, address, region, email
        FROM public.deliveries 
        WHERE order_uid = $1`, orderUID).
		Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email,
		)
	if err != nil {
		return nil, fmt.Errorf("select delivery failed: %w", err)
	}

	// 3. Получаем информацию об оплате
	err = tx.QueryRow(ctx, `
        SELECT 
            transaction, request_id, currency, provider, amount, 
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM public.payments 
        WHERE order_uid = $1`, orderUID).
		Scan(
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
			&order.Payment.CustomFee,
		)
	if err != nil {
		return nil, fmt.Errorf("select payment failed: %w", err)
	}

	// 4. Получаем товары в заказе
	rows, err := tx.Query(ctx, `
        SELECT 
            chrt_id, track_number, price, rid, name, 
            sale, size, total_price, nm_id, brand, status
        FROM public.items 
        WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("select items failed: %w", err)
	}
	defer rows.Close()

	order.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		err = rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NMID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item failed: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	return &order, nil
}

// Получить все заказы из БД
func (r *Repository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	var orders []models.Order

	rows, err := tx.Query(ctx, `
		SELECT 
            order_uid, track_number, entry, locale, 
            internal_signature, customer_id, delivery_service, 
            shardkey, sm_id, date_created, oof_shard
        FROM public.orders `)

	if err != nil {
		return nil, fmt.Errorf("scan of all orders failed: %w", err)
	}

	for rows.Next() {
		var order models.Order
		if err = rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
		); err != nil {
			return nil, fmt.Errorf("scan order failed: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	return orders, nil
}
