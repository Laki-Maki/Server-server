/*
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"subscription-service/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id string) (*model.Subscription, error)
	List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id string) error
	AggregateTotal(ctx context.Context, from, to string, userID, serviceName *string) (int, error)
}

type store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) Repository {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, sub *model.Subscription) error {
	log.Printf("[Create] serviceName=%s price=%d userID=%s startDate=%s endDate=%v",
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Format("2006-01-02"), sub.EndDate)

	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := s.db.QueryRowContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).Scan(&sub.ID); err != nil {
		log.Printf("[Create] failed: %v", err)
		return err
	}

	log.Printf("[Create] success id=%s", sub.ID)
	return nil
}

func (s *store) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	log.Printf("[GetByID] id=%s", id)

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1
	`
	var sub model.Subscription
	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		log.Printf("[GetByID] failed: %v", err)
		return nil, err
	}

	log.Printf("[GetByID] success id=%s serviceName=%s", sub.ID, sub.ServiceName)
	return &sub, nil
}

func (s *store) List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error) {
	log.Printf("[List] userID=%s serviceName=%s limit=%d offset=%d", userID, serviceName, limit, offset)

	qb := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
	`
	conds := []string{}
	args := []interface{}{}
	argIdx := 1
	if userID != "" {
		conds = append(conds, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}
	if serviceName != "" {
		conds = append(conds, fmt.Sprintf("service_name ILIKE $%d", argIdx))
		args = append(args, "%"+serviceName+"%")
		argIdx++
	}
	if len(conds) > 0 {
		qb += " WHERE " + strings.Join(conds, " AND ")
	}
	qb += fmt.Sprintf(" ORDER BY start_date DESC LIMIT %d OFFSET %d", limit, offset)

	rows := []*model.Subscription{}
	if err := s.db.SelectContext(ctx, &rows, qb, args...); err != nil {
		log.Printf("[List] failed: %v", err)
		return nil, err
	}

	log.Printf("[List] success count=%d", len(rows))
	return rows, nil
}

func (s *store) Update(ctx context.Context, sub *model.Subscription) error {
	log.Printf("[Update] id=%s serviceName=%s price=%d", sub.ID, sub.ServiceName, sub.Price)

	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
	`
	res, err := s.db.ExecContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID,
	)
	if err != nil {
		log.Printf("[Update] failed: %v", err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		log.Printf("[Update] failed to get RowsAffected: %v", err)
		return err
	}
	if n == 0 {
		log.Printf("[Update] no rows updated for id=%s", sub.ID)
		return sql.ErrNoRows
	}

	log.Printf("[Update] success id=%s updatedRows=%d", sub.ID, n)
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	log.Printf("[Delete] id=%s", id)

	res, err := s.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		log.Printf("[Delete] failed: %v", err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		log.Printf("[Delete] failed to get RowsAffected: %v", err)
		return err
	}
	if n == 0 {
		log.Printf("[Delete] no rows deleted for id=%s", id)
		return sql.ErrNoRows
	}

	log.Printf("[Delete] success id=%s deletedRows=%d", id, n)
	return nil
}

func (s *store) AggregateTotal(ctx context.Context, from, to string, userID, serviceName *string) (int, error) {
	log.Printf("[AggregateTotal] from=%s to=%s userID=%v serviceName=%v", from, to, userID, serviceName)

	query := `
SELECT COALESCE(SUM(price * months), 0)::bigint AS total FROM (
  SELECT price,
    CASE
      WHEN LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')) >= GREATEST(start_date, to_date($1,'MM-YYYY')) THEN
        (
          (date_part('year', age(LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')), GREATEST(start_date, to_date($1,'MM-YYYY')))) * 12)
          + date_part('month', age(LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')), GREATEST(start_date, to_date($1,'MM-YYYY'))))
          + 1
        )::int
      ELSE 0
    END AS months
  FROM subscriptions
  WHERE start_date <= to_date($2,'MM-YYYY')
    AND (end_date IS NULL OR end_date >= to_date($1,'MM-YYYY'))
    AND ($3::uuid IS NULL OR user_id = $3::uuid)
    AND ($4::text IS NULL OR service_name ILIKE $4::text)
) t;
`

	var uid interface{} = nil
	var sname interface{} = nil
	if userID != nil && *userID != "" {
		uid = *userID
	}
	if serviceName != nil && *serviceName != "" {
		sname = "%" + *serviceName + "%"
	}

	var total int
	err := s.db.GetContext(ctx, &total, query, from, to, uid, sname)
	if err != nil {
		log.Printf("[AggregateTotal] failed: %v", err)
		return 0, err
	}

	log.Printf("[AggregateTotal] total=%d", total)
	return total, nil
}
*/

package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"subscription-service/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id string) (*model.Subscription, error)
	List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id string) error
	AggregateTotal(ctx context.Context, from, to string, userID, serviceName *string) (int64, error)
	FindSubscriptionsOverlapping(ctx context.Context, from, to string, userID, serviceName *string) ([]*model.Subscription, error)
}

type store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) Repository {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return s.db.QueryRowContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).Scan(&sub.ID)
}

func (s *store) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1
	`
	var sub model.Subscription
	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *store) List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error) {
	qb := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
	`
	conds := []string{}
	args := []interface{}{}
	argIdx := 1
	if userID != "" {
		conds = append(conds, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}
	if serviceName != "" {
		conds = append(conds, fmt.Sprintf("service_name ILIKE $%d", argIdx))
		args = append(args, "%"+serviceName+"%")
		argIdx++
	}
	if len(conds) > 0 {
		qb += " WHERE " + strings.Join(conds, " AND ")
	}
	qb += fmt.Sprintf(" ORDER BY start_date DESC LIMIT %d OFFSET %d", limit, offset)

	rows := []*model.Subscription{}
	if err := s.db.SelectContext(ctx, &rows, qb, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *store) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
	`
	res, err := s.db.ExecContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *store) AggregateTotal(ctx context.Context, from, to string, userID, serviceName *string) (int64, error) {
	query := `
SELECT COALESCE(SUM(price * months), 0)::bigint AS total FROM (
  SELECT price,
    CASE
      WHEN LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')) >= GREATEST(start_date, to_date($1,'MM-YYYY')) THEN
        (
          (date_part('year', age(LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')), GREATEST(start_date, to_date($1,'MM-YYYY')))) * 12)
          + date_part('month', age(LEAST(COALESCE(end_date, to_date($2,'MM-YYYY')), to_date($2,'MM-YYYY')), GREATEST(start_date, to_date($1,'MM-YYYY'))))
          + 1
        )::int
      ELSE 0
    END AS months
  FROM subscriptions
  WHERE start_date <= to_date($2,'MM-YYYY')
    AND (end_date IS NULL OR end_date >= to_date($1,'MM-YYYY'))
    AND ($3::uuid IS NULL OR user_id = $3::uuid)
    AND ($4::text IS NULL OR service_name ILIKE $4::text)
) t;
`
	var uid interface{} = nil
	var sname interface{} = nil
	if userID != nil && *userID != "" {
		uid = *userID
	}
	if serviceName != nil && *serviceName != "" {
		sname = "%" + *serviceName + "%"
	}

	var total int64
	if err := s.db.GetContext(ctx, &total, query, from, to, uid, sname); err != nil {
		return 0, err
	}
	return total, nil
}

func (s *store) FindSubscriptionsOverlapping(ctx context.Context, from, to string, userID, serviceName *string) ([]*model.Subscription, error) {
	query := `
    SELECT id, service_name, price, user_id, start_date, end_date
    FROM subscriptions
    WHERE start_date <= to_date($2,'MM-YYYY')
    AND (end_date IS NULL OR end_date >= to_date($1,'MM-YYYY'))
    AND ($3::uuid IS NULL OR user_id = $3::uuid)
    AND ($4::text IS NULL OR service_name ILIKE $4::text)
    ORDER BY service_name
    `

	var uid interface{} = nil
	var sname interface{} = nil
	if userID != nil && *userID != "" {
		uid = *userID
	}
	if serviceName != nil && *serviceName != "" {
		sname = "%" + *serviceName + "%"
	}

	subs := []*model.Subscription{}
	if err := s.db.SelectContext(ctx, &subs, query, from, to, uid, sname); err != nil {
		return nil, err
	}
	return subs, nil
}
