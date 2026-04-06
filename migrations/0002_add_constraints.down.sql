ALTER TABLE subscriptions
  DROP CONSTRAINT IF EXISTS subscriptions_end_date_not_before_start_date;

ALTER TABLE subscriptions
  DROP CONSTRAINT IF EXISTS subscriptions_price_non_negative;

