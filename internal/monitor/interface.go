package monitor

import (
	"github.com/meshplus/bitxhub-model/pb"
)

//go:generate mockgen -destination mock_monitor/mock_monitor.go -package mock_monitor -source interface.go
type Monitor interface {
	// Start starts the service of monitor
	Start() error
	// Stop stops the service of monitor
	Stop() error
	// listen on interchain ibtp from appchain
	ListenIBTP() <-chan *pb.IBTP
	// listen on to-update meta info from appchain
	ListenUpdateMeta() <-chan *pb.UpdateMeta
	// listen on to-mint asset exchange event from appchain
	ListenLockEvent() <-chan *pb.LockEvent
	// query historical ibtp by its id
	QueryIBTP(id string) (*pb.IBTP, error)
	// QueryLatestMeta queries latest index map of ibtps threw on appchain
	QueryOuterMeta() map[string]uint64
}
