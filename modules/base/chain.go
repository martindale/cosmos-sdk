package base

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameChain = "chain"
)

// Chain enforces that this tx was bound to the named chain
type Chain struct {
	stack.PassInitState
	stack.PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (Chain) Name() string {
	return NameChain
}

var _ stack.Middleware = Chain{}

// CheckTx makes sure we are on the proper chain - fulfills Middlware interface
func (c Chain) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	stx, err := c.checkChainTx(ctx.ChainID(), ctx.BlockHeight(), tx)
	if err != nil {
		return res, err
	}
	return next.CheckTx(ctx, store, stx)
}

// DeliverTx makes sure we are on the proper chain - fulfills Middlware interface
func (c Chain) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	stx, err := c.checkChainTx(ctx.ChainID(), ctx.BlockHeight(), tx)
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ctx, store, stx)
}

// checkChainTx makes sure the tx is a Chain Tx, it is on the proper chain,
// and it has not expired.
func (c Chain) checkChainTx(chainID string, height int64, tx sdk.Tx) (sdk.Tx, error) {
	// make sure it is a chaintx
	ctx, ok := tx.Unwrap().(ChainTx)
	if !ok {
		return tx, ErrNoChain()
	}

	// basic validation
	err := ctx.ValidateBasic()
	if err != nil {
		return tx, err
	}

	// compare against state
	if ctx.ChainID != chainID {
		return tx, ErrWrongChain(ctx.ChainID)
	}
	if ctx.ExpiresAt != 0 && ctx.ExpiresAt <= height {
		return tx, ErrExpired()
	}
	return ctx.Tx, nil
}
