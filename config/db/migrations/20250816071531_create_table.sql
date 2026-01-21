-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    deleted_at timestamptz,
    parent_id uuid REFERENCES categories(id) ON DELETE SET NULL ON UPDATE CASCADE,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_categories_type ON categories(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);
CREATE INDEX idx_categories_type_parent ON categories(type, parent_id) WHERE deleted_at IS NULL;

CREATE TABLE transactions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    deleted_at timestamptz,
    wallet_id uuid NOT NULL,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    amount numeric(18,2) NOT NULL,
    transaction_date timestamptz zone NOT NULL,
    description text
);

CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_category_id ON transactions(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_date ON transactions(transaction_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_deleted_at ON transactions(deleted_at);
CREATE INDEX idx_transactions_wallet_date ON transactions(wallet_id, transaction_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_category_date ON transactions(category_id, transaction_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_wallet_category ON transactions(wallet_id, category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_updated_at ON transactions(updated_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE attachments (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    deleted_at timestamptz,
    transaction_id uuid NOT NULL REFERENCES transactions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    image text,
    format text,
    size bigint
);

CREATE INDEX idx_attachments_transaction_id ON attachments(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_attachments_deleted_at ON attachments(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attachments CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
-- +goose StatementEnd
