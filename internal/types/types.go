package types

import (
	"database/sql"
)

type PeerInfo struct {
	Peer string
	Endpoint string
	LatestHandshake string
	Transfer string
	Start string
	End sql.NullString
}

type PeerInfoWithLatestEndpoint struct {
	PeerInfo
	LatestEndpoint string
}