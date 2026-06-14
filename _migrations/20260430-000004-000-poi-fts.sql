CREATE VIRTUAL TABLE IF NOT EXISTS poi_fts USING fts5(
    name,
    display_name,
    category,
    content='poi_cache',
    content_rowid='rowid',
    tokenize='porter unicode61'
);
