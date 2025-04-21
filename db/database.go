package db

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func InitDB(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ensure database is closed on error
	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	createTable := `
	CREATE TABLE IF NOT EXISTS snapshot_files (
		original_path TEXT PRIMARY KEY,
		md5 TEXT NOT NULL,
		permission TEXT,
		snapshot_id INTEGER,
		remote_hash TEXT,
		status TEXT DEFAULT 'pending'
	);
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		status TEXT DEFAULT 'pending'
	);`

	if _, err = db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func SaveFileInfo(db *sql.DB, originalPath, md5, permission string, snapshotID int, remoteHash, status string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if er := tx.Rollback(); er != nil {
				panic(er)
			}
			panic(p)
		} else if err != nil {
			if er := tx.Rollback(); er != nil {
				panic(er)
			}
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO snapshot_files 
		(original_path, md5, permission, snapshot_id, remote_hash, status) 
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(originalPath, md5, permission, snapshotID, remoteHash, status)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type FileRecord struct {
	Path       string
	MD5        string
	Permission string
	SnapId     int
	RemoteHash string
	Status     string
}

func ListFiles(db *sql.DB) ([]FileRecord, error) {
	rows, err := db.Query(`SELECT original_path, md5, permission, snapshot_id, remote_hash, status FROM snapshot_files ORDER BY original_path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileRecord

	for rows.Next() {
		var f FileRecord
		// var modTimeStr string
		if err := rows.Scan(&f.Path, &f.MD5, &f.Permission, &f.SnapId, &f.RemoteHash, &f.Status); err != nil {
			return nil, err
		}
		// f.Modified, _ = time.Parse(time.RFC3339, modTimeStr)
		files = append(files, f)
	}
	return files, nil
}

func GetFileByHash(db *sql.DB, hash string) (*FileRecord, error) {
	var f FileRecord
	query := `SELECT original_path, md5, permission, snapshot_id, remote_hash, status FROM snapshot_files WHERE md5 = ?`
	err := db.QueryRow(query, hash).Scan(&f.Path, &f.MD5, &f.Permission, &f.SnapId, &f.RemoteHash, &f.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No matching record found
		}
		return nil, err
	}
	return &f, nil
}

func ListFilesbySnapshot(db *sql.DB, snapshot_id int) ([]FileRecord, error) {
	rows, err := db.Query(`SELECT original_path, md5, permission, snapshot_id, remote_hash,status FROM snapshot_files WHERE snapshot_id = ? ORDER BY original_path`, snapshot_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []FileRecord
	for rows.Next() {
		var f FileRecord
		if err := rows.Scan(&f.Path, &f.MD5, &f.Permission, &f.SnapId, &f.RemoteHash, &f.Status); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func SaveSnapshot(db *sql.DB) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO snapshots (date) VALUES (datetime('now'))")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec()
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

type SnapShotRecord struct {
	Id     int
	Date   string
	Status string
}

func ListSnapShots(db *sql.DB) ([]SnapShotRecord, error) {
	rows, err := db.Query(`SELECT id, date, status FROM snapshots ORDER BY date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snaps []SnapShotRecord

	for rows.Next() {
		var f SnapShotRecord
		// var modTimeStr string
		if err := rows.Scan(&f.Id, &f.Date, &f.Status); err != nil {
			return nil, err
		}
		// f.Modified, _ = time.Parse(time.RFC3339, modTimeStr)
		snaps = append(snaps, f)
	}
	return snaps, nil
}

func GetSnapByDate(db *sql.DB, date string) (*SnapShotRecord, error) {
	var f SnapShotRecord
	query := `SELECT id, date, status FROM snapshots WHERE date = ?`
	err := db.QueryRow(query, date).Scan(&f.Id, &f.Date, &f.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No matching record found
		}
		return nil, err
	}
	return &f, nil
}

func GetLastSnap(db *sql.DB) (*SnapShotRecord, error) {
	var f SnapShotRecord
	query := `SELECT id, date, status FROM snapshots ORDER BY date DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&f.Id, &f.Date, &f.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No matching record found
		}
		return nil, err
	}
	return &f, nil
}

func UpdateSnapshotFileField(db *sql.DB, recordID int, fieldName string, newValue any) error {
	// Validate field name to prevent SQL injection
	validFields := map[string]bool{
		"original_path": true,
		"md5":           true,
		"permission":    true,
		"snapshot_id":   true,
		"remote_hash":   true,
		"status":        true,
	}
	if !validFields[fieldName] {
		return fmt.Errorf("invalid field name: %s", fieldName)
	}

	query := fmt.Sprintf("UPDATE snapshot_files SET %s = ? WHERE id = ?", fieldName)
	_, err := db.Exec(query, newValue, recordID)
	if err != nil {
		return fmt.Errorf("failed to update field: %w", err)
	}

	return nil
}
