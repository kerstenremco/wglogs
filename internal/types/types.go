package types

type PeerInfo struct {
	Peer string
	Endpoint string
	LatestHandshake string
	Transfer string
	Start string
	End string
}

type PeerInfoWithLatestEndpoint struct {
	PeerInfo
	LatestEndpoint string
}