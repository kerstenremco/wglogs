package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kerstenremco/wglogs/internal/db"
	"github.com/kerstenremco/wglogs/internal/types"
	"github.com/kerstenremco/wglogs/internal/wg"
)


	
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command: sync or sync-test")
		return
	}
	firstArg := os.Args[1]
	db.CreateDatabase()

	switch firstArg {
	case "svc":
		service()
	case "sync":
		sync(false)
	case "sync-test":
		sync(true)
	case "show":
		conn := db.OpenDatabase()
		defer conn.Close()
		entries := db.GetAllEntries(conn)
		printTable(entries)
	default:
		log.Fatal("Unknown argument")
	}
	
}

func printTable(entries []types.PeerInfo) {
	t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Peer", "Endpoint", "Latest Handshake", "Transfer", "Start", "End"})
		for _, entry := range entries {
			t.AppendRow(table.Row{
				entry.Peer[0:10] + "...",
				entry.Endpoint,
				strings.NewReplacer("minutes", "min", "seconds ago", "sec").Replace(entry.LatestHandshake),
				strings.NewReplacer("received", "recv").Replace(entry.Transfer),
				entry.Start,
				entry.End,
			})
		}
		t.Render()
}

func sync(testMode bool) {
	conn := db.OpenDatabase()
	defer conn.Close()

	// Get current peer info
	results, err := wg.GetInfo(testMode)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resultsExtended := db.GetLatestEndpoint(conn, results)

	// Prepare entries to insert, update, close
	var entriesToClose []types.PeerInfoWithLatestEndpoint
	var entriesToInsert []types.PeerInfoWithLatestEndpoint
	var entriesToUpdate []types.PeerInfoWithLatestEndpoint

	for _, entry := range resultsExtended {
		// Close old entries if endpoint has changed
		if entry.Endpoint != entry.LatestEndpoint && entry.LatestEndpoint != "" {
			entriesToClose = append(entriesToClose, entry)
		}
		// Insert new entry if endpoint is different
		if entry.Endpoint != entry.LatestEndpoint {
			entriesToInsert = append(entriesToInsert, entry)
		} else {
			entriesToUpdate = append(entriesToUpdate, entry)
		}
	}

	// Execute database changes
	db.CloseEntry(conn, entriesToClose)
	db.InsertEntries(conn, entriesToInsert)
	db.UpdateEntry(conn, entriesToUpdate)
}

func service() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("Service started...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sync(false)
	fmt.Println("Sync completed at", time.Now().Format(time.RFC1123))

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Service stopping...")
			return
		case <-ticker.C:
			sync(false)
			fmt.Println("Sync completed at", time.Now().Format(time.RFC1123))
		}
	}
}