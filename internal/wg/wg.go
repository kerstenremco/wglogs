package wg

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/kerstenremco/wglogs/internal/types"
)

func GetInfo(testMode bool) ([]types.PeerInfo, error) {
	var results []types.PeerInfo

	var cmd *exec.Cmd
	if testMode {
		cmd = exec.Command("cat", "example.txt")
	} else {
		cmd = exec.Command("wg")
	}
	out, err := cmd.Output()

	if err != nil {
		return results, errors.New("failed to execute wg command")
	}

		scanner := bufio.NewScanner(bytes.NewReader(out))
		var peerInfo types.PeerInfo = types.PeerInfo{}
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "peer:") {
				parts := strings.SplitN(line, ": ", 2)
				peerInfo.Peer = parts[1]
			}
			if strings.Contains(line, "endpoint:") {
				parts := strings.SplitN(line, ": ", 2)
				peerInfo.Endpoint = parts[1]
			}
			if strings.Contains(line, "latest handshake:") {
				parts := strings.SplitN(line, ": ", 2)
				peerInfo.LatestHandshake = parts[1]
			}
			if strings.Contains(line, "transfer:") {
				parts := strings.SplitN(line, ": ", 2)
				peerInfo.Transfer = parts[1]

				// Finished reading one peer's info
				results = append(results, peerInfo)
				peerInfo = types.PeerInfo{}
			}
	}
	return results, nil
}