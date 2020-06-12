package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	mail "github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.uber.org/zap"
)

var Email = &email{
	Transfer: t{},
}

type email struct {
	Transfer t
}

type t struct{}

func (transfer *t) Initiate(req *types.TransferReq) {
	mail.Transfer.Initiate(req)
}

func (transfer *t) Accept(j *types.Journal) {
	info, err := transfer.getTransferEmailInfo(j)
	if err != nil {
		l.Logger.Error("logic.Email.Transfer.Accept failed", zap.Error(err))
		return
	}
	mail.Transfer.Accept(info)
}

func (transfer *t) Reject(j *types.Journal, reason string) {
	info, err := transfer.getTransferEmailInfo(j, reason)
	if err != nil {
		l.Logger.Error("logic.Email.Transfer.Reject failed", zap.Error(err))
		return
	}
	mail.Transfer.Reject(info)
}

func (transfer *t) Cancel(j *types.Journal, reason string) {
	info, err := transfer.getTransferEmailInfo(j, reason)
	if err != nil {
		l.Logger.Error("logic.Email.Transfer.Cancel failed", zap.Error(err))
		return
	}
	mail.Transfer.Cancel(info)
}

func (transfer *t) CancelBySystem(j *types.Journal, reason string) {
	info, err := transfer.getTransferEmailInfo(j, reason)
	if err != nil {
		l.Logger.Error("logic.Email.Transfer.CancelBySystem failed", zap.Error(err))
		return
	}
	mail.Transfer.CancelBySystem(info)
}

func (transfer *t) getTransferEmailInfo(j *types.Journal, reason ...string) (*mail.TransferEmailInfo, error) {
	info := &mail.TransferEmailInfo{
		Amount: j.Amount,
	}
	if len(reason) > 0 {
		info.Reason = reason[0]
	}

	fromEntity, err := Entity.FindByAccountNumber(j.FromAccountNumber)
	if err != nil {
		return nil, err
	}
	toEntity, err := Entity.FindByAccountNumber(j.ToAccountNumber)
	if err != nil {
		return nil, err
	}

	if j.InitiatedBy == j.FromAccountNumber {
		info.TransferDirection = "out"
		info.InitiatorEmail = fromEntity.Email
		info.InitiatorEntityName = fromEntity.Name
		info.ReceiverEmail = toEntity.Email
		info.ReceiverEntityName = toEntity.Name
	} else {
		info.TransferDirection = "in"
		info.InitiatorEmail = toEntity.Email
		info.InitiatorEntityName = toEntity.Name
		info.ReceiverEmail = fromEntity.Email
		info.ReceiverEntityName = fromEntity.Name
	}

	return info, nil
}
