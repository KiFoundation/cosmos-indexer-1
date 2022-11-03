package ibc

import (
	"fmt"
	parsingTypes "github.com/DefiantLabs/cosmos-tax-cli-private/cosmos/modules"

	types "github.com/DefiantLabs/cosmos-tax-cli-private/cosmos/modules/ibc/types"
	txModule "github.com/DefiantLabs/cosmos-tax-cli-private/cosmos/modules/tx"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
)

const (
	MsgTransfer = "/ibc.applications.transfer.v1.MsgTransfer"
)

type WrapperMsgTransfer struct {
	txModule.Message
	CosmosMsgTransfer *types.MsgTransfer
	SenderAddress     string
	ReceiverAddress   string
	Amount            *stdTypes.Coin
}

// HandleMsg: Handle type checking for MsgFundCommunityPool
func (sf *WrapperMsgTransfer) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgTransfer = msg.(*types.MsgTransfer)

	//Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	//Funds sent and sender address are pulled from the parsed Cosmos Msg
	sf.SenderAddress = sf.CosmosMsgTransfer.Sender
	sf.ReceiverAddress = sf.CosmosMsgTransfer.Receiver
	sf.Amount = &sf.CosmosMsgTransfer.Token

	return nil
}

func (sf *WrapperMsgTransfer) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	if sf.Amount != nil {
		return []parsingTypes.MessageRelevantInformation{{
			SenderAddress:        sf.SenderAddress,
			ReceiverAddress:      sf.ReceiverAddress,
			AmountSent:           sf.Amount.Amount.BigInt(),
			AmountReceived:       sf.Amount.Amount.BigInt(),
			DenominationSent:     sf.Amount.Denom,
			DenominationReceived: sf.Amount.Denom,
		}}
	}
	return nil
}

func (sf *WrapperMsgTransfer) String() string {
	if sf.Amount == nil {
		return fmt.Sprintf("MsgTransfer: IBC transfer from %s to %s did not include an amount\n", sf.SenderAddress, sf.ReceiverAddress)
	}
	return fmt.Sprintf("MsgTransfer: IBC transfer of %s from %s to %s\n", sf.CosmosMsgTransfer.Token, sf.SenderAddress, sf.ReceiverAddress)
}