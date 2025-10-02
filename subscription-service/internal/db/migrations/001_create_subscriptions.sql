-- internal/db/migrations/001_create_subscriptions.sql
CREATE TABLE IF NOT EXISTS subscriptions (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
service_name TEXT NOT NULL,
price INTEGER NOT NULL,
user_id UUID NOT NULL,
start_date DATE NOT NULL,
end_date DATE
);


CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service ON subscriptions(service_name);
