-- +goose Up
-- +goose StatementBegin
Create TABLE department {
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT | NULL,
    created_at DATE
};
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
