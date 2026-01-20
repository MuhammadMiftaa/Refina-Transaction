-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW view_category_group_by_type AS
SELECT 
	parent.name AS group_name,
	JSON_AGG(
		JSON_BUILD_OBJECT(
			'id', child.id,
			'name', child.name
		)
		ORDER BY child.name
	) AS category,
	parent.type AS type
FROM categories parent
LEFT JOIN categories child ON child.parent_id = parent.id
WHERE parent.parent_id IS NULL AND parent.deleted_at IS NULL
GROUP BY parent.name, parent.type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS view_category_group_by_type;
-- +goose StatementEnd
