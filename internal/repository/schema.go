package repository

import "database/sql"

func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		source_path TEXT NOT NULL,
		source_type TEXT NOT NULL DEFAULT 'folder',
		schedule TEXT NOT NULL,
		enabled BOOLEAN NOT NULL,
		created_at DATETIME NOT NULL
	);


	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		status TEXT NOT NULL CHECK(status IN ('OK','ERROR')),
		size_bytes INTEGER NOT NULL,
		source_path TEXT NOT NULL,
		source_type TEXT NOT NULL,
		target_path TEXT NOT NULL,
		error_message TEXT,
		checksum TEXT,
		started_at DATETIME NOT NULL,
		finished_at DATETIME NOT NULL,
		FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_backups_task_id 
	ON backups(task_id);

	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),

		backup_root_path TEXT NOT NULL,
		max_storage_bytes INTEGER NOT NULL,
		theme_mode TEXT NOT NULL
	);

	INSERT OR IGNORE INTO settings (
		backup_root_path,
		max_storage_bytes,
		theme_mode
	) VALUES (
		'',
		0,
		'system'
	);
	`

	_, err := db.Exec(schema)
	return err
}
