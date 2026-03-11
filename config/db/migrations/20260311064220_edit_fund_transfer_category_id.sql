-- +goose Up
-- +goose StatementBegin
INSERT INTO categories (id, parent_id, name, type) VALUES
('00000000-0000-0000-0000-000000000010', NULL,                                   'Pindah Dana', 'fund_transfer'),
('00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000010', 'Cash In',     'fund_transfer'),
('00000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000010', 'Cash Out',    'fund_transfer')
ON CONFLICT (id) DO UPDATE SET
  parent_id = EXCLUDED.parent_id,
  name      = EXCLUDED.name,
  type      = EXCLUDED.type;

UPDATE transactions
SET category_id = '00000000-0000-0000-0000-000000000011'
FROM categories
WHERE transactions.category_id = categories.id
  AND categories.name = 'Cash In';

UPDATE transactions
SET category_id = '00000000-0000-0000-0000-000000000012'
FROM categories
WHERE transactions.category_id = categories.id
  AND categories.name = 'Cash Out';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE transactions
SET category_id = '75a19b5c-43b4-4f84-ad2c-f844c40eec24'
FROM categories
WHERE transactions.category_id = categories.id
  AND categories.name = 'Cash In';

UPDATE transactions
SET category_id = 'b5a5d097-6346-4b72-b975-8ebf0b4b72f1'
FROM categories
WHERE transactions.category_id = categories.id
  AND categories.name = 'Cash Out';

DELETE FROM categories WHERE id IN (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000012'
);

INSERT INTO categories (id, parent_id, name, type) VALUES
('e8f83876-0e5c-4e88-b16b-f3229d5c8412', NULL,                                      'Pindah Dana', 'fund_transfer'),
('75a19b5c-43b4-4f84-ad2c-f844c40eec24', 'e8f83876-0e5c-4e88-b16b-f3229d5c8412',   'Cash In',     'fund_transfer'),
('b5a5d097-6346-4b72-b975-8ebf0b4b72f1', 'e8f83876-0e5c-4e88-b16b-f3229d5c8412',   'Cash Out',    'fund_transfer')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd