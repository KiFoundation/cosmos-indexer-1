package main

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

// Works on cosmos bby default. If you need to use this for other chains you must set the prefix.
//
// See comments in address.go in cosmos-sdk:
//	config := sdk.GetConfig()
//	config.SetBech32PrefixForAccount(yourBech32PrefixAccAddr, yourBech32PrefixAccPub)
//	config.SetBech32PrefixForValidator(yourBech32PrefixValAddr, yourBech32PrefixValPub)
//	config.SetBech32PrefixForConsensusNode(yourBech32PrefixConsAddr, yourBech32PrefixConsPub)
func TestCosmosHubAddressEquality(t *testing.T) {
	valoperAddress := "cosmosvaloper130mdu9a0etmeuw52qfxk73pn0ga6gawkxsrlwf" //strangelove's valoper
	accountAddress := "cosmos130mdu9a0etmeuw52qfxk73pn0ga6gawkryh2z6"        //strangelove's delegator
	cosmAccountAddress, acctErr := types.AccAddressFromBech32(accountAddress)
	cosmValAccountAddress, valoperErr := types.ValAddressFromBech32(valoperAddress)

	if acctErr != nil || valoperErr != nil || !cosmAccountAddress.Equals(cosmValAccountAddress) {
		t.Fatal("Addresses not equivalent", acctErr, valoperErr)
	}
}

//Works on all chains but you need to know the prefix (e.g. junovaloper) in advance
func TestCosmosAllAddressEquality(t *testing.T) {
	valoperAddress := "junovaloper130mdu9a0etmeuw52qfxk73pn0ga6gawk2tz77l" //strangelove's valoper
	accountAddress := "juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x"        //strangelove's delegator
	cosmAccountAddress, acctErr := types.GetFromBech32(accountAddress, "juno")
	cosmValAccountAddress, valoperErr := types.GetFromBech32(valoperAddress, "junovaloper")

	if acctErr != nil || valoperErr != nil || !bytes.Equal(cosmAccountAddress, cosmValAccountAddress) {
		t.Fatal("Addresses not equivalent", acctErr, valoperErr)
	}

	junovaloperAddr := types.MustBech32ifyAddressBytes("junovaloper", cosmAccountAddress)
	if junovaloperAddr != valoperAddress {
		t.Fatal("Addresses not equivalent", junovaloperAddr)
	}
}
