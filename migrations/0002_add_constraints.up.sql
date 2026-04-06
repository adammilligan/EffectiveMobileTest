ALTER TABLE subscriptions
  ADD CONSTRAINT subscriptions_price_non_negative CHECK (price >= 0);

ALTER TABLE subscriptions
  ADD CONSTRAINT subscriptions_end_date_not_before_start_date
  CHECK (end_date IS NULL OR end_date >= start_date);

