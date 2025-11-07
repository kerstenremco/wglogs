package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/kerstenremco/wglogs/internal/types"
	_ "modernc.org/sqlite"
)

func OpenDatabase() (*sql.DB) {
	db, err := sql.Open("sqlite", "wglogs.db")
	if err != nil {
		log.Fatal("failed to open database")
	}
	return db
}

func GetAllEntries(db *sql.DB) ([]types.PeerInfo) {
	var results []types.PeerInfo

	querySQL := `SELECT peer, endpoint, latest_handshake, transfer, start, end FROM peers ORDER BY start DESC`
	rows, err := db.Query(querySQL)
	if err != nil {
		log.Fatal("failed to query data")
	}
	defer rows.Close()

	for rows.Next() {
		var entry types.PeerInfo
		err := rows.Scan(&entry.Peer, &entry.Endpoint, &entry.LatestHandshake, &entry.Transfer, &entry.Start, &entry.End)
		if err != nil {
			log.Fatal("failed to scan data")
		}
		results = append(results, entry)
	}
	return results
}

func InsertEntries(db *sql.DB, rows []types.PeerInfoWithLatestEndpoint) {
	insertSQL := `INSERT INTO peers (peer, endpoint, latest_handshake, transfer, start, end) VALUES (?, ?, ?, ?, ?, ?)`
	for _, result := range rows {
		_, err := db.Exec(insertSQL, result.Peer, result.Endpoint, result.LatestHandshake, result.Transfer, result.Start, "")
		if err != nil {
			log.Fatal("failed to insert data")
		}
	}
}

func UpdateEntry(db *sql.DB, rows []types.PeerInfoWithLatestEndpoint) {
	for _, row := range rows {

		updateSQL := `UPDATE peers SET transfer = ?, latest_handshake = ? WHERE id = (
			SELECT id FROM peers WHERE peer = ? ORDER BY id DESC LIMIT 1
		)`
		_, err := db.Exec(updateSQL, row.Transfer, row.LatestHandshake, row.Peer)
		if err != nil {
			log.Fatal("failed to update entry")
		}
	}
}

func CloseEntry(db *sql.DB, rows []types.PeerInfoWithLatestEndpoint) {
	
	current := time.Now()
	for _, row := range rows {
		updateSQL := `UPDATE peers SET end = ? WHERE id = (
			SELECT id FROM peers WHERE peer = ? AND endpoint = ? ORDER BY id DESC LIMIT 1
		)`
		_, err := db.Exec(updateSQL, current.Format("2006-01-02 15:04:05"), row.Peer, row.LatestEndpoint)
		if err != nil {
			log.Fatal("failed to close entry")
		}
	}
}

func GetLatestEndpoint(db *sql.DB, entries []types.PeerInfo) ([]types.PeerInfoWithLatestEndpoint) {
	var results []types.PeerInfoWithLatestEndpoint

	for _, entry := range entries {
		querySQL := `SELECT endpoint FROM peers WHERE peer = ? ORDER BY id DESC LIMIT 1`
		row := db.QueryRow(querySQL, entry.Peer)
		
		var latestEndpoint string
		err := row.Scan(&latestEndpoint)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal("failed to get latest endpoint")
		} else {
			results = append(results, types.PeerInfoWithLatestEndpoint{PeerInfo: entry, LatestEndpoint: latestEndpoint})
		}
	}
	return results
}

func CreateDatabase() {
	conn := OpenDatabase()
	defer conn.Close()
	createTableSQL := `CREATE TABLE IF NOT EXISTS peers (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"peer" TEXT NOT NULL,
		"endpoint" TEXT NOT NULL,
		"latest_handshake" TEXT NOT NULL,
		"transfer" TEXT NOT NULL,
		"start" TEXT NOT NULL,
		"end" TEXT NOT NULL
	  );`

	_, err := conn.Exec(createTableSQL)

	if err != nil {
		log.Fatal("Error creating database")
	}
}