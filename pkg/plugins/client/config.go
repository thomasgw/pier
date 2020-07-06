package client

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/meshplus/pier/pkg/plugins/proto"
	"google.golang.org/grpc"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  3,
	MagicCookieKey:   "PIER_APPCHAIN_PLUGIN",
	MagicCookieValue: "PIER",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"fabric-plugin": &AppchainGRPCPlugin{},
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type AppchainGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Client
}

func (p *AppchainGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterAppchainPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *AppchainGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client:      proto.NewAppchainPluginClient(c),
		doneContect: ctx,
	}, nil
}
