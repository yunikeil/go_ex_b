WITH seed_users AS (
    INSERT INTO users (email, username, password_hash)
    VALUES
        ('alice@example.com', 'alice_demo', crypt('password123', gen_salt('bf', 10))),
        ('bob@example.com', 'bob_demo', crypt('password123', gen_salt('bf', 10)))
    ON CONFLICT (email) DO UPDATE
        SET username = EXCLUDED.username
    RETURNING id, email
),
all_seed_users AS (
    SELECT id, email FROM seed_users
    UNION
    SELECT id, email FROM users WHERE email IN ('alice@example.com', 'bob@example.com')
),
seed_accounts AS (
    INSERT INTO accounts (user_id, number, currency, balance)
    SELECT id, number, 'RUB', balance
    FROM (
        SELECT id, '40817810000000000001' AS number, 150000.00 AS balance
        FROM all_seed_users WHERE email = 'alice@example.com'
        UNION ALL
        SELECT id, '40817810000000000002' AS number, 75000.00 AS balance
        FROM all_seed_users WHERE email = 'bob@example.com'
    ) data
    ON CONFLICT (number) DO UPDATE
        SET balance = EXCLUDED.balance
    RETURNING id, user_id, number
),
all_seed_accounts AS (
    SELECT id, user_id, number FROM seed_accounts
    UNION
    SELECT id, user_id, number
    FROM accounts
    WHERE number IN ('40817810000000000001', '40817810000000000002')
)
INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description, created_at)
SELECT user_id, NULL::BIGINT, id, 'deposit', 150000.00, 'seed opening balance', now() - interval '10 days'
FROM all_seed_accounts
WHERE number = '40817810000000000001'
  AND NOT EXISTS (
      SELECT 1 FROM transactions
      WHERE to_account_id = all_seed_accounts.id AND description = 'seed opening balance'
  )
UNION ALL
SELECT user_id, NULL::BIGINT, id, 'deposit', 75000.00, 'seed opening balance', now() - interval '9 days'
FROM all_seed_accounts
WHERE number = '40817810000000000002'
  AND NOT EXISTS (
      SELECT 1 FROM transactions
      WHERE to_account_id = all_seed_accounts.id AND description = 'seed opening balance'
  );

INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description, created_at)
SELECT alice.user_id, alice.id, bob.id, 'transfer', 2500.00, 'seed transfer to bob', now() - interval '3 days'
FROM accounts alice
JOIN accounts bob ON bob.number = '40817810000000000002'
WHERE alice.number = '40817810000000000001'
  AND NOT EXISTS (
      SELECT 1 FROM transactions
      WHERE from_account_id = alice.id AND to_account_id = bob.id AND description = 'seed transfer to bob'
  );

INSERT INTO cards (user_id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, last4)
SELECT
    a.user_id,
    a.id,
    pgp_sym_encrypt('2202123412341237', 'dev-card-pgp-key-change-me'),
    pgp_sym_encrypt('12/29', 'dev-card-pgp-key-change-me'),
    crypt('123', gen_salt('bf', 10)),
    encode(hmac('2202123412341237|12/29', 'dev-card-hmac-key-change-me', 'sha256'), 'hex'),
    '1237'
FROM accounts a
WHERE a.number = '40817810000000000001'
  AND NOT EXISTS (
      SELECT 1 FROM cards WHERE account_id = a.id AND last4 = '1237'
  );

WITH alice_account AS (
    SELECT a.id AS account_id, a.user_id
    FROM accounts a
    JOIN users u ON u.id = a.user_id
    WHERE u.email = 'alice@example.com' AND a.number = '40817810000000000001'
),
seed_credit AS (
    INSERT INTO credits (user_id, account_id, principal, annual_rate, term_months, monthly_payment, status, next_payment_at)
    SELECT user_id, account_id, 120000.00, 21.00, 12, 11172.61, 'active', now() + interval '1 month'
    FROM alice_account
    WHERE NOT EXISTS (
        SELECT 1 FROM credits c
        WHERE c.account_id = alice_account.account_id
          AND c.principal = 120000.00
          AND c.term_months = 12
    )
    RETURNING id
),
all_seed_credit AS (
    SELECT id FROM seed_credit
    UNION
    SELECT c.id
    FROM credits c
    JOIN alice_account a ON a.account_id = c.account_id
    WHERE c.principal = 120000.00 AND c.term_months = 12
)
INSERT INTO payment_schedules (credit_id, due_date, amount)
SELECT credit_id, due_date, 11172.61
FROM (
    SELECT id AS credit_id, now() + (n || ' months')::interval AS due_date
    FROM all_seed_credit
    CROSS JOIN generate_series(1, 12) AS n
) schedule
WHERE NOT EXISTS (
    SELECT 1 FROM payment_schedules ps
    WHERE ps.credit_id = schedule.credit_id AND ps.due_date::date = schedule.due_date::date
);
