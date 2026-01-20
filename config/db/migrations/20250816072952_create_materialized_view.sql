-- +goose Up

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS view_user_wallet_daily_summaries AS
WITH date_series AS (
SELECT generate_series(
	CURRENT_DATE - INTERVAL '89 days',
	CURRENT_DATE,
	INTERVAL '1 day'
)::date AS date
),
wallet_info AS (
SELECT
	w.id AS wallet_id,
	w.user_id,
	wt.type AS wallet_type,
	w.balance AS current_balance
FROM wallets w
JOIN wallet_types wt ON wt.id = w.wallet_type_id
),
tx_summary AS (
SELECT
	t.wallet_id,
	DATE(t.transaction_date) AS date,
	SUM(
	CASE
		WHEN c.type = 'expense' OR (c.type = 'fund_transfer' AND c.name = 'Cash Out') THEN -1 * t.amount
		WHEN c.type = 'income' OR (c.type = 'fund_transfer' AND c.name = 'Cash In') THEN t.amount
		ELSE 0
	END
	) AS amount
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.deleted_at IS NULL
GROUP BY t.wallet_id, DATE(t.transaction_date)
),
daily_reverse_cumulative AS (
SELECT
	ds.date,
	wi.user_id,
	wi.wallet_id,
	wi.wallet_type,
	wi.current_balance 
	- COALESCE((
		SELECT SUM(ts2.amount)
		FROM tx_summary ts2
		WHERE ts2.wallet_id = wi.wallet_id
		AND ts2.date > ds.date
	), 0) AS total_amount
FROM date_series ds
CROSS JOIN wallet_info wi
),
pivoted AS (
SELECT
	date,
	user_id,
	SUM(CASE WHEN wallet_type = 'physical' THEN total_amount ELSE 0 END) AS physical,
	SUM(CASE WHEN wallet_type = 'e-wallet' THEN total_amount ELSE 0 END) AS e_wallet,
	SUM(CASE WHEN wallet_type = 'bank' THEN total_amount ELSE 0 END) AS bank,
	SUM(CASE WHEN wallet_type NOT IN ('physical', 'e-wallet', 'bank') THEN total_amount ELSE 0 END) AS others
FROM daily_reverse_cumulative
GROUP BY date, user_id
)
SELECT * FROM pivoted
ORDER BY user_id, date;

CREATE INDEX IF NOT EXISTS idx_view_user_wallet_daily_summaries_user_id ON view_user_wallet_daily_summaries (user_id, date);

SELECT cron.schedule('0 18 * * *', 'REFRESH MATERIALIZED VIEW view_user_wallet_daily_summaries');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS view_user_summaries AS		
WITH
current_month AS (
SELECT
	u.id AS user_id,
	SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END) AS income_now,
	SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END) AS expense_now
FROM users u
JOIN wallets w ON w.user_id = u.id
LEFT JOIN transactions t ON t.wallet_id = w.id
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_date >= date_trunc('month', current_date)
	AND t.transaction_date < date_trunc('month', current_date + INTERVAL '1 month')
GROUP BY u.id
),
previous_month AS (
SELECT
	u.id AS user_id,
	SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END) AS income_prev,
	SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END) AS expense_prev
FROM users u
JOIN wallets w ON w.user_id = u.id
LEFT JOIN transactions t ON t.wallet_id = w.id
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_date >= date_trunc('month', current_date - INTERVAL '1 month')
	AND t.transaction_date < date_trunc('month', current_date)
GROUP BY u.id
),
current_balance AS (
SELECT
	u.id AS user_id,
	SUM(w.balance) AS balance_now
FROM users u
JOIN wallets w ON w.user_id = u.id
GROUP BY u.id
),
previous_balance AS (
SELECT
	user_id,
	(physical + e_wallet + bank + others) AS balance_prev
FROM view_user_wallet_daily_summaries
WHERE date = (date_trunc('month', current_date) - INTERVAL '1 day')::date
)
SELECT
u.id AS user_id,
u.name,
COALESCE(cm.income_now, 0) AS income_now,
COALESCE(cm.expense_now, 0) AS expense_now,
COALESCE(cm.income_now, 0) - COALESCE(cm.expense_now, 0) AS profit_now,
COALESCE(cb.balance_now, 0) AS balance_now,
ROUND((
	(COALESCE(cm.income_now, 0) - COALESCE(pm.income_prev, 0)) /
	NULLIF(pm.income_prev, 0)
) * 100, 2) AS user_income_growth_percentage,
ROUND((
	(COALESCE(cm.expense_now, 0) - COALESCE(pm.expense_prev, 0)) /
	NULLIF(pm.expense_prev, 0)
) * 100, 2) AS user_expense_growth_percentage,
ROUND((
	((COALESCE(cm.income_now, 0) - COALESCE(cm.expense_now, 0)) -
	(COALESCE(pm.income_prev, 0) - COALESCE(pm.expense_prev, 0))) /
	NULLIF((COALESCE(pm.income_prev, 0) - COALESCE(pm.expense_prev, 0)), 0)
) * 100, 2) AS user_profit_growth_percentage,
ROUND((
	(COALESCE(cb.balance_now, 0) - COALESCE(pb.balance_prev, 0)) /
	NULLIF(pb.balance_prev, 0)
) * 100, 2) AS user_balance_growth_percentage
FROM users u
LEFT JOIN current_month cm ON cm.user_id = u.id
LEFT JOIN previous_month pm ON pm.user_id = u.id
LEFT JOIN current_balance cb ON cb.user_id = u.id
LEFT JOIN previous_balance pb ON pb.user_id = u.id;

CREATE INDEX IF NOT EXISTS idx_view_user_summaries_user_id ON view_user_summaries (user_id);

SELECT cron.schedule('0 18 * * *', 'REFRESH MATERIALIZED VIEW view_user_summaries');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS view_user_monthly_summaries AS
SELECT 
	u.id AS user_id,
	to_char(t.transaction_date, 'YYYY-MM') AS month,
	rtrim(to_char(t.transaction_date, 'month')) as month_name,
	SUM(CASE WHEN c."type" = 'income' THEN t.amount ELSE 0 END) AS total_income,
	SUM(CASE WHEN c."type" = 'expense' THEN t.amount ELSE 0 END) AS total_expense
FROM 
	users u
JOIN 
	wallets w ON u.id = w.user_id
JOIN 
	transactions t ON w.id = t.wallet_id
JOIN
	categories c ON c.id = t.category_id
WHERE 
	t.transaction_date >= date_trunc('month', CURRENT_DATE) - INTERVAL '11 months'
GROUP BY 
	u.id, to_char(t.transaction_date, 'YYYY-MM'), to_char(t.transaction_date, 'month')
ORDER BY 
	u.id, to_char(t.transaction_date, 'YYYY-MM') ASC;

CREATE INDEX IF NOT EXISTS idx_view_user_monthly_summaries_user_id ON view_user_monthly_summaries (user_id, month_name);

SELECT cron.schedule('0 18 * * *', 'REFRESH MATERIALIZED VIEW view_user_monthly_summaries');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS view_user_most_expenses AS
WITH transaction_totals AS (
	SELECT 
		user_id,
		parent.name AS parent_category_name,
		SUM(transactions.amount) AS total,
		ROW_NUMBER() OVER (
		PARTITION BY user_id 
		ORDER BY SUM(transactions.amount) DESC
		) AS rank
	FROM users
	LEFT JOIN wallets ON wallets.user_id = users.id
	LEFT JOIN transactions ON transactions.wallet_id = wallets.id
	LEFT JOIN categories ON categories.id = transactions.category_id
	LEFT JOIN categories parent ON parent.id = categories.parent_id
	WHERE categories."type" = 'expense'
		AND transactions.transaction_date >= date_trunc('month', CURRENT_DATE) - INTERVAL '2 months'
	GROUP BY parent.name, user_id
)
SELECT *
FROM transaction_totals
WHERE rank <= 7
ORDER BY user_id ASC, total DESC;

CREATE INDEX IF NOT EXISTS idx_view_user_most_expenses_user_id ON view_user_most_expenses (user_id, parent_category_name);

SELECT cron.schedule('0 18 * * *', 'REFRESH MATERIALIZED VIEW view_user_most_expenses');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM cron.job WHERE command LIKE 'REFRESH MATERIALIZED VIEW view_user_summaries%';
DROP INDEX IF EXISTS idx_view_user_summaries_user_id;
DROP MATERIALIZED VIEW IF EXISTS view_user_summaries;
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM cron.job WHERE command LIKE 'REFRESH MATERIALIZED VIEW view_user_monthly_summaries%';
DROP INDEX IF EXISTS idx_view_user_monthly_summaries_user_id;
DROP MATERIALIZED VIEW IF EXISTS view_user_monthly_summaries;
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM cron.job WHERE command LIKE 'REFRESH MATERIALIZED VIEW view_user_most_expenses%';
DROP INDEX IF EXISTS idx_view_user_most_expenses_user_id;
DROP MATERIALIZED VIEW IF EXISTS view_user_most_expenses;
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM cron.job WHERE command LIKE 'REFRESH MATERIALIZED VIEW view_user_wallet_daily_summaries%';
DROP INDEX IF EXISTS idx_view_user_wallet_daily_summaries_user_id;
DROP MATERIALIZED VIEW IF EXISTS view_user_wallet_daily_summaries;
-- +goose StatementEnd
