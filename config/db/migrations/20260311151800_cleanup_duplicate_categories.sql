-- +goose Up
-- +goose StatementBegin

-- Step 1: Point all transactions using duplicate "Deposit Awal" categories
-- to the canonical hardcoded ID (00000000-0000-0000-0000-000000000000)
UPDATE transactions
SET category_id = '00000000-0000-0000-0000-000000000000'
WHERE category_id IN (
    SELECT id FROM categories
    WHERE name = 'Deposit Awal' AND type = 'income' AND deleted_at IS NULL
      AND id != '00000000-0000-0000-0000-000000000000'
);

-- Step 2: Soft-delete duplicate "Deposit Awal" categories (keep the hardcoded one)
UPDATE categories
SET deleted_at = NOW()
WHERE name = 'Deposit Awal' AND type = 'income' AND deleted_at IS NULL
  AND id != '00000000-0000-0000-0000-000000000000';

-- Step 3: Soft-delete old "Pindah Dana" parent (keep hardcoded 00000000-0000-0000-0000-000000000010)
UPDATE categories
SET deleted_at = NOW()
WHERE name = 'Pindah Dana' AND type = 'fund_transfer' AND deleted_at IS NULL
  AND id != '00000000-0000-0000-0000-000000000010';

-- Step 4: Soft-delete old "Cash In" (keep hardcoded 00000000-0000-0000-0000-000000000011)
UPDATE categories
SET deleted_at = NOW()
WHERE name = 'Cash In' AND type = 'fund_transfer' AND deleted_at IS NULL
  AND id != '00000000-0000-0000-0000-000000000011';

-- Step 5: Soft-delete old "Cash Out" (keep hardcoded 00000000-0000-0000-0000-000000000012)
UPDATE categories
SET deleted_at = NOW()
WHERE name = 'Cash Out' AND type = 'fund_transfer' AND deleted_at IS NULL
  AND id != '00000000-0000-0000-0000-000000000012';

-- Step 6: Soft-delete duplicate "Cicilan Kendaraan" categories
-- Keep the one with the seeded parent_id (86453a2c-...)
UPDATE categories
SET deleted_at = NOW()
WHERE name = 'Cicilan Kendaraan' AND type = 'expense' AND deleted_at IS NULL
  AND id != '78662c66-1299-4486-902c-7ea8543aa6fc';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore soft-deleted duplicates (undo cleanup)
UPDATE categories
SET deleted_at = NULL
WHERE deleted_at IS NOT NULL
  AND name IN ('Deposit Awal', 'Pindah Dana', 'Cash In', 'Cash Out', 'Cicilan Kendaraan')
  AND id NOT IN (
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000012',
    '78662c66-1299-4486-902c-7ea8543aa6fc'
  );

-- +goose StatementEnd
