package state

import "database/sql"

// SetCache stores a key-value pair in the cache table.
func (s *Store) SetCache(key, value string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO cache (key, value) VALUES (?, ?)`,
		key, value,
	)
	return err
}

// GetCache retrieves a value from the cache table.
func (s *Store) GetCache(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM cache WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}
