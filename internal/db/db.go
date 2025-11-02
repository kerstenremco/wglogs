package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kerstenremco/wglogs/internal/types"

	_ "modernc.org/sqlite"
)

func GetAllEntries() ([]types.PeerInfo, error) {
	var results []types.PeerInfo
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return results, errors.New("failed to open database")
	}
	defer db.Close()

	querySQL := `SELECT peer, endpoint, latest_handshake, transfer, start, end FROM peers`
	rows, err := db.Query(querySQL)
	if err != nil {
		return results, errors.New("failed to query data")
	}
	defer rows.Close()

	for rows.Next() {
		var entry types.PeerInfo
		err := rows.Scan(&entry.Peer, &entry.Endpoint, &entry.LatestHandshake, &entry.Transfer, &entry.Start, &entry.End)
		if err != nil {
			return results, errors.New("failed to scan data")
		}
		results = append(results, entry)
	}
	return results, nil
}

func InsertEntries(rows []types.PeerInfoWithLatestEndpoint) error {
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return errors.New("failed to open database")
	}
	defer db.Close()

	insertSQL := `INSERT INTO peers (peer, endpoint, latest_handshake, transfer) VALUES (?, ?, ?, ?)`
	for _, result := range rows {
		_, err = db.Exec(insertSQL, result.Peer, result.Endpoint, result.LatestHandshake, result.Transfer)
		if err != nil {
			return errors.New("failed to insert data")
		}
	}
	return nil
}

func UpdateEntry(rows []types.PeerInfoWithLatestEndpoint) error {
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return errors.New("failed to open database")
	}
	defer db.Close()

	for _, row := range rows {

		updateSQL := `UPDATE peers SET transfer = ?, latest_handshake = ? WHERE id = (
			SELECT id FROM peers WHERE peer = ? ORDER BY id DESC LIMIT 1
		)`
		_, err = db.Exec(updateSQL, row.Transfer, row.LatestHandshake, row.Peer)
		if err != nil {
			return errors.New("failed to update entry")
		}
	}
	return nil
}

func CloseEntry(rows []types.PeerInfoWithLatestEndpoint) error {
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return errors.New("failed to open database")
	}
	defer db.Close()

	for _, row := range rows {

		updateSQL := `UPDATE peers SET end = datetime('now') WHERE id = (
			SELECT id FROM peers WHERE peer = ? AND end IS NULL ORDER BY id DESC LIMIT 1
		)`
		_, err = db.Exec(updateSQL, row.Peer)
		if err != nil {
			return errors.New("failed to close entry")
		}
	}
	return nil
}

func GetLatestEndpoint(entries []types.PeerInfo) ([]types.PeerInfoWithLatestEndpoint, error) {
	var results []types.PeerInfoWithLatestEndpoint
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return nil, errors.New("failed to open database")
	}
	defer db.Close()

	for _, entry := range entries {
		querySQL := `SELECT endpoint FROM peers WHERE peer = ? ORDER BY id DESC LIMIT 1`
		row := db.QueryRow(querySQL, entry.Peer)
		var latestEndpoint string
		err := row.Scan(&latestEndpoint)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println("Error querying data:", err)
		} else {
			results = append(results, types.PeerInfoWithLatestEndpoint{PeerInfo: entry, LatestEndpoint: latestEndpoint})
		}
	}
	return results, nil
}

func CreateDatabase() error {
	db, err := sql.Open("sqlite", "./wglogs.db")
	if err != nil {
		return errors.New("failed to open database")
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS peers (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"peer" TEXT NOT NULL,
		"endpoint" TEXT NOT NULL,
		"latest_handshake" TEXT NOT NULL,
		"transfer" TEXT NOT NULL,
		"start" DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
		"end" DATETIME DEFAULT NULL
	  );`
	  
	_, err = db.Exec(createTableSQL)

	if err != nil {
		return errors.New("failed to create table")
	}
	return nil
}