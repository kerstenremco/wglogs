package main

import (
	"context"
	"fmt"
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

	switch firstArg {
	case "svc":
		service()
	case "sync":
		sync(false)
	case "sync-test":
		sync(true)
	case "show":
		entries, err := db.GetAllEntries()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		printTable(entries)
	default:
		fmt.Println("Unknown command:", firstArg)
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
				entry.End.String,
			})
		}
		t.Render()
}

func sync(testMode bool) {
	err := db.CreateDatabase()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Get current peer info
	results, err := wg.GetInfo(testMode)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resultsExtended, err := db.GetLatestEndpoint(results)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

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
	db.CloseEntry(entriesToClose)
	db.InsertEntries(entriesToInsert)
	db.UpdateEntry(entriesToUpdate)
}

func service() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("Service started...")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	sync(true)
	fmt.Println("Sync completed at", time.Now().Format(time.RFC1123))

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Service stopping...")
			return
		case <-ticker.C:
			sync(true)
			fmt.Println("Sync completed at", time.Now().Format(time.RFC1123))
		}
	}
}