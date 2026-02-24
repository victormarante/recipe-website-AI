-- Create recipes table with JSON columns for arrays
CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    categories TEXT NOT NULL,   -- JSON array: ["breakfast", "vegetarian"]
    ingredients TEXT NOT NULL,  -- JSON array: ["1 cup flour", "2 eggs"]
    steps TEXT NOT NULL,        -- JSON array: ["Mix ingredients", "Cook"]
    links TEXT,                 -- JSON array: [{"type": "external", "url": "...", "label": "..."}]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index for search performance
CREATE INDEX IF NOT EXISTS idx_recipes_title ON recipes(title);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_recipes_updated_at
    AFTER UPDATE ON recipes
    FOR EACH ROW
BEGIN
    UPDATE recipes SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
