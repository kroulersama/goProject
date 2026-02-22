-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_departments_parent 
        FOREIGN KEY (parent_id) 
        REFERENCES departments(id) 
        ON DELETE CASCADE
);

-- Индекс для поиска по parent_id
CREATE INDEX idx_departments_parent_id ON departments(parent_id);
CREATE UNIQUE INDEX idx_departments_name_parent 
ON departments(name, COALESCE(parent_id, 0));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd