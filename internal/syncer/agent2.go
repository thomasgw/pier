package syncer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/meshplus/pier/internal/utils"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	appchainmgr "github.com/meshplus/bitxhub-core/appchain-mgr"
	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub-model/constant"
	"github.com/meshplus/bitxhub-model/pb"
	rpcx "github.com/meshplus/go-bitxhub-client"
	"github.com/meshplus/pier/internal/repo"
	"github.com/meshplus/pier/pkg/model"
	"github.com/sirupsen/logrus"
)

func (syncer *WrapperSyncer2) GetAssetExchangeSigns(id string) ([]byte, error) {
	resp, err := utils.GetMultiSigns(syncer.peerMgr, id, pb.GetMultiSignsRequest_ASSET_EXCHANGE)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Sign == nil {
		return nil, fmt.Errorf("get empty signatures for asset exchange id %s", id)
	}

	var signs []byte
	for _, sign := range resp.Sign {
		signs = append(signs, sign...)
	}

	return signs, nil
}

func (syncer *WrapperSyncer2) GetIBTPSigns(ibtp *pb.IBTP) ([]byte, error) {
	hash := ibtp.Hash()
	resp, err := utils.GetMultiSigns(syncer.peerMgr, hash.String(), pb.GetMultiSignsRequest_IBTP)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Sign == nil {
		return nil, fmt.Errorf("get empty signatures for ibtp %s", ibtp.ID())
	}
	signs, err := resp.Marshal()
	if err != nil {
		return nil, err
	}

	return signs, nil
}

func (syncer *WrapperSyncer2) GetAppchains() ([]*appchainmgr.Appchain, error) {
	var receipt *pb.Receipt
	if err := syncer.retryFunc(func(attempt uint) error {
		tx, err := utils.GenerateContractTx(syncer.priv, pb.TransactionData_BVM, constant.AppchainMgrContractAddr.Address(), "Appchains")
		if err != nil {
			return err
		}
		tx.Nonce = 1
		err = tx.Sign(syncer.priv)
		if err != nil {
			return fmt.Errorf("%w: for reason %s", utils.ErrSignTx, err.Error())
		}
		receipt, err = utils.SenView(syncer.peerMgr, tx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}

	ret := make([]*appchainmgr.Appchain, 0)
	if receipt == nil || receipt.Ret == nil {
		return ret, nil
	}
	if err := json.Unmarshal(receipt.Ret, &ret); err != nil {
		return nil, err
	}
	appchains := make([]*appchainmgr.Appchain, 0)
	for _, appchain := range ret {
		if appchain.ChainType != repo.BitxhubType {
			appchains = append(appchains, appchain)
		}
	}
	return appchains, nil
}

func (syncer *WrapperSyncer2) GetInterchainById(from string) *pb.Interchain {
	ic := &pb.Interchain{}
	var receipt *pb.Receipt
	tx, err := utils.GenerateContractTx(syncer.priv, pb.TransactionData_BVM, constant.InterchainContractAddr.Address(), "GetInterchain", rpcx.String(from))
	if err != nil {
		syncer.logger.Errorf("GetInterchainById generateContractTx err:%s", err)
		return ic
	}
	tx.Nonce = 1
	err = tx.Sign(syncer.priv)
	if err != nil {
		syncer.logger.Errorf("GetInterchainById sign err:%s", err)
		return ic
	}
	receipt, err = utils.SenView(syncer.peerMgr, tx)
	if err != nil {
		syncer.logger.Errorf("GetInterchainById SenView err:%s", err)
		return ic
	}
	var interchain pb.Interchain
	if err := interchain.Unmarshal(receipt.Ret); err != nil {
		return ic
	}
	return &interchain
}

func (syncer *WrapperSyncer2) QueryInterchainMeta() *pb.Interchain {
	var interchainMeta *pb.Interchain
	var receipt *pb.Receipt
	if err := syncer.retryFunc(func(attempt uint) error {
		queryTx, err := utils.GenerateContractTx(syncer.priv, pb.TransactionData_BVM,
			constant.InterchainContractAddr.Address(), "Interchain")
		if err != nil {
			return err
		}
		queryTx.Nonce = 1
		err = queryTx.Sign(syncer.priv)
		if err != nil {
			return fmt.Errorf("%w: for reason %s", utils.ErrSignTx, err.Error())
		}
		receipt, err = utils.SenView(syncer.peerMgr, queryTx)
		if err != nil {
			return err
		}
		if !receipt.IsSuccess() {
			return fmt.Errorf("receipt: %s", receipt.Ret)
		}
		ret := &pb.Interchain{}
		if err := ret.Unmarshal(receipt.Ret); err != nil {
			return fmt.Errorf("unmarshal interchain meta from bitxhub: %w", err)
		}
		interchainMeta = ret
		return nil
	}); err != nil {
		syncer.logger.Panicf("query interchain meta: %s", err.Error())
	}

	return interchainMeta
}

func (syncer *WrapperSyncer2) QueryIBTP(ibtpID string) (*pb.IBTP, bool, error) {
	var receipt *pb.Receipt
	queryTx, err := utils.GenerateContractTx(syncer.priv, pb.TransactionData_BVM, constant.InterchainContractAddr.Address(),
		"GetIBTPByID", rpcx.String(ibtpID))
	if err != nil {
		return nil, false, err
	}
	queryTx.Nonce = 1
	err = queryTx.Sign(syncer.priv)
	if err != nil {
		return nil, false, fmt.Errorf("%w: for reason %s", utils.ErrSignTx, err.Error())
	}
	receipt, err = utils.SenView(syncer.peerMgr, queryTx)
	if err != nil {
		return nil, false, err
	}

	if !receipt.IsSuccess() {
		return nil, false, fmt.Errorf("%w: %s", ErrIBTPNotFound, string(receipt.Ret))
	}

	hash := types.NewHash(receipt.Ret)

	response, err := utils.GetTransaction(syncer.peerMgr, hash.String())
	if err != nil {
		return nil, false, err
	}
	receipt, err = utils.GetReceipt(syncer.peerMgr, hash.String())
	if err != nil {
		return nil, false, err
	}
	return response.Txs.Transactions[0].GetIBTP(), receipt.Status == pb.Receipt_SUCCESS, nil
}

func (syncer *WrapperSyncer2) ListenIBTP() <-chan *model.WrappedIBTP {
	return syncer.ibtpC
}

func (syncer *WrapperSyncer2) SendIBTP(ibtp *pb.IBTP) error {
	proof := ibtp.GetProof()
	proofHash := sha256.Sum256(proof)
	ibtp.Proof = proofHash[:]

	tx, _ := utils.GenerateIBTPTx(syncer.priv, ibtp)
	tx.Extra = proof

	var receipt *pb.Receipt
	err := syncer.retryFunc(func(attempt uint) error {
		hash, err := utils.SendTransaction(syncer.priv, syncer.peerMgr, tx, nil)
		if err != nil {
			syncer.logger.Errorf("Send ibtp error: %s", err.Error())
			if errors.Is(err, rpcx.ErrReconstruct) {
				tx, _ = utils.GenerateIBTPTx(syncer.priv, ibtp)
				tx.Extra = proof
			}
			return err
		}
		syncer.logger.WithFields(logrus.Fields{"hash": hash}).Info("syncer agent sendIBTP to bxh")
		receipt, err = utils.GetReceipt(syncer.peerMgr, hash)
		syncer.logger.WithFields(logrus.Fields{"receiptHash": receipt.TxHash, "status": receipt.Status}).Info("get receipt")
		if err != nil {
			return fmt.Errorf("get tx receipt by hash %s: %w", hash, err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	if !receipt.IsSuccess() {
		syncer.logger.WithFields(logrus.Fields{
			"ibtp_id": ibtp.ID(),
			"msg":     string(receipt.Ret),
		}).Error("Receipt result for ibtp")
		// if no rule bind for this appchain or appchain not available, exit pier
		errMsg := string(receipt.Ret)
		if strings.Contains(errMsg, noBindRule) ||
			strings.Contains(errMsg, srcchainNotAvailable) {
			return fmt.Errorf("appchain not valid: %s", errMsg)
		}
		// if target chain is not available, this ibtp should be rollback
		if strings.Contains(errMsg, dstchainNotAvailable) {
			syncer.logger.Errorf("Destination appchain is not available: %s, try to rollback in source appchain...", string(receipt.Ret))
			syncer.rollbackHandler(ibtp)
			return nil
		}
		if strings.Contains(errMsg, ibtpIndexExist) {
			// if ibtp index is lower than index recorded on bitxhub, then ignore this ibtp
			return nil
		}
		if strings.Contains(errMsg, ibtpIndexWrong) {
			// if index is wrong ,notify exchanger to update its meta from bitxhub
			return ErrMetaOutOfDate
		}
		if strings.Contains(errMsg, invalidIBTP) {
			// if this ibtp structure is not compatible or verify failed
			// try to get new ibtp and resend
			return fmt.Errorf("invalid ibtp %s", ibtp.ID())
		}
		return fmt.Errorf("unknown error, retry for %s anyway", ibtp.ID())
	}
	syncer.logger.WithFields(logrus.Fields{
		"ibtp_id": ibtp.ID(),
		"msg":     string(receipt.Ret),
	}).Info("Receipt result for bixthub")
	return nil
}

func (syncer *WrapperSyncer2) retryFunc(handle func(uint) error) error {
	return retry.Retry(func(attempt uint) error {
		if err := handle(attempt); err != nil {
			syncer.logger.Errorf("retry failed for reason: %s", err.Error())
			return err
		}
		return nil
	}, strategy.Wait(500*time.Millisecond))
}