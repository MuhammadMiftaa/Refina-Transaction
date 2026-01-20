-- +goose Up
-- +goose StatementBegin
INSERT INTO investment_types (id, name, unit) VALUES
('192676ca-569f-4e73-b2fc-0d1e772006f4', 'Others',	                '-'),
('d90a40cb-6981-4847-b54b-a0affdcaf446', 'Gold',	                'Gram / Troy Ounce'),
('a0e5075e-0fab-46ab-9719-4e3c97ba6ce5', 'Government Securities',   'Nominal / Unit'),
('c1991cd9-6e96-49cb-877b-c9fc3ea71303', 'Bonds',	                'Nominal / Lot'),
('75d32c45-a81a-4ac1-a331-c4b7c22e2e75', 'Stocks',	                'Lembar'),
('e18f9601-5331-4aa4-a0bc-960a7f46261d', 'Mutual Funds',            'Unit Penyertaan (UP)'),
('46113757-d86d-49b7-b604-a670ca5b40b1', 'Deposits',	            'Nominal');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE TABLE investment_types;
-- +goose StatementEnd
