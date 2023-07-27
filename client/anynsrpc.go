package anynsclient

import (
	"context"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/pool"
	"github.com/anyproto/any-sync/net/rpc/rpcerr"
	"github.com/anyproto/any-sync/nodeconf"

	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
)

/*
 * This client should be used to access the Anytype Naming Service (Anyns)
 */
type AnyNsClientService interface {
	IsNameAvailable(ctx context.Context, in *as.NameAvailableRequest) (out *as.NameAvailableResponse, err error)
	GetOperationStatus(ctx context.Context, in *as.GetOperationStatusRequest) (out *as.OperationResponse, err error)
	NameRegisterSigned(ctx context.Context, in *as.NameRegisterSignedRequest) (out *as.OperationResponse, err error)

	app.ComponentRunnable
}

type service struct {
	pool     pool.Pool
	nodeconf nodeconf.Service
	close    chan struct{}
}

const CName = "any-ns.anynsclient"

var log = logger.NewNamed(CName)

func (s *service) Init(a *app.App) (err error) {
	s.pool = a.MustComponent(pool.CName).(pool.Pool)
	s.nodeconf = a.MustComponent(nodeconf.CName).(nodeconf.Service)
	s.close = make(chan struct{})
	return nil
}

func (s *service) Name() (name string) {
	return CName
}

func New() AnyNsClientService {
	return new(service)
}

func (s *service) Run(_ context.Context) error {
	return nil
}

func (s *service) Close(_ context.Context) error {
	select {
	case <-s.close:
	default:
		close(s.close)
	}
	return nil
}

func (s *service) doClient(ctx context.Context, fn func(cl as.DRPCAnynsClient) error) error {
	// TODO: check here...
	peer, err := s.pool.Get(ctx, s.nodeconf.ConsensusPeers()[0])
	if err != nil {
		return err
	}

	dc, err := peer.AcquireDrpcConn(ctx)
	if err != nil {
		return err
	}
	defer peer.ReleaseDrpcConn(dc)

	return fn(as.NewDRPCAnynsClient(dc))
}

func (s *service) IsNameAvailable(ctx context.Context, in *as.NameAvailableRequest) (out *as.NameAvailableResponse, err error) {
	err = s.doClient(ctx, func(cl as.DRPCAnynsClient) error {
		if out, err = cl.IsNameAvailable(ctx, in); err != nil {
			return rpcerr.Unwrap(err)
		}
		return nil
	})
	return
}

func (s *service) GetOperationStatus(ctx context.Context, in *as.GetOperationStatusRequest) (out *as.OperationResponse, err error) {
	err = s.doClient(ctx, func(cl as.DRPCAnynsClient) error {
		if out, err = cl.GetOperationStatus(ctx, in); err != nil {
			return rpcerr.Unwrap(err)
		}
		return nil
	})
	return
}

// unisgned NameRegister method is not implemented! However, server still has it
func (s *service) NameRegisterSigned(ctx context.Context, in *as.NameRegisterSignedRequest) (out *as.OperationResponse, err error) {
	err = s.doClient(ctx, func(cl as.DRPCAnynsClient) error {
		if out, err = cl.NameRegisterSigned(ctx, in); err != nil {
			return rpcerr.Unwrap(err)
		}
		return nil
	})
	return
}
