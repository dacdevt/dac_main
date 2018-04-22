package dac

import (
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/dacchain/dacapp/common"
	"github.com/dacchain/dacapp/miner"
)

// PublicdacPI provides an API to access dacChain full node-related
// information.
type PublicdacAPI struct {
	e *dac
}

// NewPublicdacAPI creates a new dacChain protocol API for full nodes.
func NewPublicdacAPI(e *dac) *PublicdacAPI {
	return &PublicdacAPI{e}
}

// dacrbase is the address that mining rewards will be send to
func (api *PublicdacAPI) dacbase() (common.Address, error) {
	return api.e.dacbase()
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (api *PublicdacAPI) Coinbase() (common.Address, error) {
	return api.dacbase()
}

// PrivateAdminAPI is the collection of dacChain full node-related APIs
// exposed over the private admin endpoint.
type PrivateAdminAPI struct {
	dac *dac
}

// NewPrivateAdminAPI creates a new API definition for the full node private
// admin methods of the dacChain service.
func NewPrivateAdminAPI(dac *dac) *PrivateAdminAPI {
	return &PrivateAdminAPI{dac: dac}
}

// ExportChain exports the current blockchain into a local file.
func (api *PrivateAdminAPI) ExportChain(file string) (bool, error) {
	// Make sure we can create the file to export into
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false, err
	}
	defer out.Close()

	var writer io.Writer = out
	if strings.HasSuffix(file, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}

	// Export the blockchain
	if err := api.dac.BlockChain().Export(writer); err != nil {
		return false, err
	}
	return true, nil
}

// PublicMinerAPI provides an API to control the miner.
// It offers only methods that operate on data that pose no security risk when it is publicly accessible.
type PublicMinerAPI struct {
	e     *dac
	agent *miner.RemoteAgent
}

// NewPublicMinerAPI create a new PublicMinerAPI instance.
func NewPublicMinerAPI(e *dac) *PublicMinerAPI {
	agent := miner.NewRemoteAgent(e.BlockChain(), e.Engine())
	e.Miner().Register(agent)

	return &PublicMinerAPI{e, agent}
}

// Mining returns an indication if this node is currently mining.
func (api *PublicMinerAPI) Mining() bool {
	return api.e.IsMining()
}
