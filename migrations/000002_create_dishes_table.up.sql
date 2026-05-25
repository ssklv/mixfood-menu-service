CREATE TABLE IF NOT EXISTS dishes (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    weight INT,
    volume NUMERIC(10, 2),
    proteins NUMERIC(10, 2) NOT NULL DEFAULT 0,
    fats NUMERIC(10, 2) NOT NULL DEFAULT 0,
    carbs NUMERIC(10, 2) NOT NULL DEFAULT 0,
    calories INT NOT NULL DEFAULT 0,
    image_url TEXT NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_dishes_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
);