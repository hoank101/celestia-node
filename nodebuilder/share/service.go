package share

import (
	"context"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/share/service"

	"github.com/ipfs/go-blockservice"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt/namespace"
)

// Module provides access to any data square or block share on the network.
//
// All Get methods provided on Module follow the following flow:
//  1. Check local storage for the requested Share.
//  2. If exists
//     * Load from disk
//     * Return
//  3. If not
//     * Find provider on the network
//     * Fetch the Share from the provider
//     * Store the Share
//     * Return
//
// Any method signature changed here needs to also be changed in the API struct.
//
//go:generate mockgen -destination=mocks/api.go -package=mocks . Module
type Module interface {
	// SharesAvailable subjectively validates if Shares committed to the given Root are available on the Network.
	SharesAvailable(context.Context, *share.Root) error
	// ProbabilityOfAvailability calculates the probability of the data square
	// being available based on the number of samples collected.
	ProbabilityOfAvailability() float64
	GetShare(ctx context.Context, dah *share.Root, row, col int) (share.Share, error)
	GetShares(ctx context.Context, root *share.Root) ([][]share.Share, error)
	// GetSharesByNamespace iterates over a square's row roots and accumulates the found shares in the given namespace.ID.
	GetSharesByNamespace(ctx context.Context, root *share.Root, namespace namespace.ID) ([]share.Share, error)
}

type module struct {
	share.Availability
	service.ShareService
}

func newModule(shrSrv service.ShareService, avail share.Availability) Module {
	return &module{
		Availability: avail,
		ShareService: shrSrv,
	}
}

func NewShareService(lc fx.Lifecycle, bServ blockservice.BlockService) service.ShareService {
	serv := service.NewShareService(bServ)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return serv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return serv.Stop(ctx)
		},
	})
	return serv
}

func NewModule(lc fx.Lifecycle, tp node.Type, shrSrv service.ShareService, availability share.Availability) Module {
	return newModule(shrSrv, availability)
}

// API is a wrapper around Module for the RPC.
// TODO(@distractedm1nd): These structs need to be autogenerated.
type API struct {
	SharesAvailable           func(context.Context, *share.Root) error
	ProbabilityOfAvailability func() float64
	GetShare                  func(ctx context.Context, dah *share.Root, row, col int) (share.Share, error)
	GetShares                 func(ctx context.Context, root *share.Root) ([][]share.Share, error)
	GetSharesByNamespace      func(ctx context.Context, root *share.Root, namespace namespace.ID) ([]share.Share, error)
}