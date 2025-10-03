package model

import "time"

type Subscription struct {
	ID          string     `db:"id" json:"id"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Price       int        `db:"price" json:"price"`
	UserID      string     `db:"user_id" json:"user_id"`
	StartDate   time.Time  `db:"start_date" json:"start_date"`
	EndDate     *time.Time `db:"end_date" json:"end_date,omitempty"`
}

type AggregateResponse struct {
	UserID        string             `json:"user_id,omitempty"`
	Subscriptions []SubscriptionInfo `json:"subscriptions"`
	From          string             `json:"from"`
	To            string             `json:"to"`
	Total         int64              `json:"total"`
}

type SubscriptionInfo struct {
	Number      int    `json:"number"` // Номер подписки в списке
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
}
