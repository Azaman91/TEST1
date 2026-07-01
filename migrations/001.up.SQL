CREATE TABLE subscriptions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    sum_price INTEGER,
    sum_price_id INTEGER,
    user_id UUID NOT NULL,
    start_date VARCHAR(7) NOT NULL,
    end_date VARCHAR(7)
)

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_start_date ON subscriptions(start_date);