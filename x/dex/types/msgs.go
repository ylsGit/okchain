package types

import (
	"github.com/cosmos/cosmos-sdk/x/auth"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgDeposit           = "deposit"
	TypeMsgWithdraw          = "withdraw"
	TypeMsgTransferOwnership = "transferOwnership"
)

type MsgList struct {
	Owner      sdk.AccAddress `json:"owner"`
	ListAsset  string         `json:"list_asset"`  //  Symbol of asset listed on Dex.
	QuoteAsset string         `json:"quote_asset"` //  Symbol of asset quoted by asset listed on Dex.
	InitPrice  sdk.Dec        `json:"init_price"`
}

func NewMsgList(owner sdk.AccAddress, listAsset, quoteAsset string, initPrice sdk.Dec) MsgList {
	return MsgList{
		Owner:      owner,
		ListAsset:  listAsset,
		QuoteAsset: quoteAsset,
		InitPrice:  initPrice,
	}
}

func (msg MsgList) Route() string { return RouterKey }

func (msg MsgList) Type() string { return "list" }

func (msg MsgList) ValidateBasic() sdk.Error {
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgList) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgList) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

type MsgDelist struct {
	Owner   sdk.AccAddress `json:"owner"`
	Product string         `json:"product"`
}

func NewMsgDelist(owner sdk.AccAddress, product string) MsgDelist {
	return MsgDelist{
		Owner:   owner,
		Product: product,
	}
}

func (msg MsgDelist) Route() string { return RouterKey }

func (msg MsgDelist) Type() string { return "delist" }

func (msg MsgDelist) ValidateBasic() sdk.Error {
	if msg.Product == "" || len(msg.Product) == 0 {
		return ErrInvalidProduct(msg.Product)
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgDelist) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgDelist) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// MsgDeposit
type MsgDeposit struct {
	Product   string         `json:"product"`   // product for trading pair in full name of the tokens
	Amount    sdk.DecCoin    `json:"amount"`    // Coins to add to the deposit
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
}

func NewMsgDeposit(product string, amount sdk.DecCoin, depositor sdk.AccAddress) MsgDeposit {
	return MsgDeposit{product, amount, depositor}
}

// Implements Msg.
// nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return TypeMsgDeposit }

// Implements Msg.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgWithdraw
type MsgWithdraw struct {
	Product   string         `json:"product"`   // product for trading pair in full name of the tokens
	Amount    sdk.DecCoin    `json:"amount"`    // Coins to add to the deposit
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
}

func NewMsgWithdraw(product string, amount sdk.DecCoin, depositor sdk.AccAddress) MsgWithdraw {
	return MsgWithdraw{product, amount, depositor}
}

// Implements Msg.
// nolint
func (msg MsgWithdraw) Route() string { return RouterKey }
func (msg MsgWithdraw) Type() string  { return TypeMsgWithdraw }

// Implements Msg.
func (msg MsgWithdraw) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

// Implements Msg.
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgTransferOwnership - high level transaction of the coin module
type MsgTransferOwnership struct {
	FromAddress sdk.AccAddress    `json:"from_address"`
	ToAddress   sdk.AccAddress    `json:"to_address"`
	Product     string            `json:"product"`
	ToSignature auth.StdSignature `json:"to_signature"`
}

func NewMsgTransferOwnership(from, to sdk.AccAddress, product string) MsgTransferOwnership {
	return MsgTransferOwnership{
		FromAddress: from,
		ToAddress:   to,
		Product:     product,
		ToSignature: auth.StdSignature{},
	}
}

// Route Implements Msg.
func (msg MsgTransferOwnership) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgTransferOwnership) Type() string { return TypeMsgTransferOwnership }

// ValidateBasic Implements Msg.
func (msg MsgTransferOwnership) ValidateBasic() sdk.Error {
	if msg.FromAddress.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}

	if msg.ToAddress.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}

	if msg.Product == "" {
		return sdk.ErrUnknownRequest("product cannot be empty")
	}

	if !msg.checkMultiSign() {
		return sdk.ErrUnauthorized("invalid multi signature")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgTransferOwnership) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

func (msg MsgTransferOwnership) checkMultiSign() bool {
	// check pubkey
	if msg.ToSignature.PubKey == nil {
		return false
	}

	if !sdk.AccAddress(msg.ToSignature.PubKey.Address()).Equals(msg.ToAddress) {
		return false
	}

	// check multisign
	toSignature := msg.ToSignature
	msg.ToSignature = auth.StdSignature{}
	toValid := toSignature.VerifyBytes(msg.GetSignBytes(), toSignature.Signature)
	return toValid
}
