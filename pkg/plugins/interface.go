package plugins

import (
	"github.com/meshplus/bitxhub-model/pb"
)

// Client defines the interface that interacts with appchain
//go:generate mockgen -destination mock_client/mock_client.go -package mock_client -source interface.go
type Client interface {
	// Initialize initialize plugin client
	Initialize(configPath string, extra []byte, mode string) error

	// Start starts to listen appchain event
	Start() error

	// Stop stops client
	Stop() error

	// GetIBTPCh gets an interchain ibtp channel generated by client
	GetIBTPCh() chan *pb.IBTP

	// GetUpdateMeta gets an updated trust meta channel by client
	GetUpdateMeta() chan *pb.UpdateMeta

	// SubmitIBTP submits the interchain ibtp to appchain
	SubmitIBTP(from string, index uint64, serviceID string, ibtpType pb.IBTP_Type, content *pb.Content, proof *pb.BxhProof, isEncrypted bool) (*pb.SubmitIBTPResponse, error)

	// SubmitIBTPBatch submit the multi interchain ibtps to appchain
	SubmitIBTPBatch(from []string, index []uint64, serviceID []string, ibtpType []pb.IBTP_Type, content []*pb.Content, proof []*pb.BxhProof, isEncrypted []bool) (*pb.SubmitIBTPResponse, error)

	// SubmitReceipt submit the multi receipt ibtp to appchain
	SubmitReceipt(to string, index uint64, serviceID string, ibtpType pb.IBTP_Type, result *pb.Result, proof *pb.BxhProof) (*pb.SubmitIBTPResponse, error)

	// SubmitReceiptBatch submit the receipt ibtp to appchain
	SubmitReceiptBatch(to []string, index []uint64, serviceID []string, ibtpType []pb.IBTP_Type, result []*pb.Result, proof []*pb.BxhProof) (*pb.SubmitIBTPResponse, error)

	// GetOutMessage gets interchain ibtp by service pair and index from broker contract
	GetOutMessage(servicePair string, idx uint64) (*pb.IBTP, error)

	// GetReceiptMessage gets receipt ibtp by service pair and index from broker contract
	GetReceiptMessage(servicePair string, idx uint64) (*pb.IBTP, error)

	// GetInMeta gets an index map, which implicates the greatest index of
	// ingoing interchain txs for each service pair
	GetInMeta() (map[string]uint64, error)

	// GetOutMeta gets an index map, which implicates the greatest index of
	// outgoing interchain txs for each service pair
	GetOutMeta() (map[string]uint64, error)

	// GetReceiptMeta gets an index map, which implicates the greatest index of
	// executed callback txs for each service pair
	GetCallbackMeta() (map[string]uint64, error)

	// GetDstRollbackMeta gets an index map, which implicates the greatest index of
	// executed rollback txs from each service pair
	GetDstRollbackMeta() (map[string]uint64, error)

	// GetDirectTransactionMeta gets transaction start timestamp, timeout period and transaction status in direct mode
	GetDirectTransactionMeta(string) (uint64, uint64, uint64, error)

	// GetServices gets all service IDs the pier cares
	GetServices() ([]string, error)

	// GetChainID gets BitXHub and appchain ID
	GetChainID() (string, string, error)

	// GetAppchainInfo gets appchain information by appchain ID
	GetAppchainInfo(chainID string) (string, []byte, string, error)

	// Name gets name of blockchain from plugin
	Name() string

	// Type gets type of blockchain from plugin
	Type() string

	// GetOffChainData get offchain data and send back
	GetOffChainData(request *pb.GetDataRequest) (*pb.GetDataResponse, error)

	// GetOffChainDataReq get offchain data request
	GetOffChainDataReq() chan *pb.GetDataRequest

	// SubmitOffChainData submit offchain data to plugin
	SubmitOffChainData(response *pb.GetDataResponse) error
}
