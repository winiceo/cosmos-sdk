package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// IBCMapper stores IBC packets under a specific store key in the multi store
// with the encoding provided by its codec.
type IBCMapper struct {
	key sdk.StoreKey
	cdc *wire.Codec
}

// NewIBCMapper returns a new IBCMapper.
func NewIBCMapper(cdc *wire.Codec, key sdk.StoreKey) IBCMapper {
	return IBCMapper{
		key: key,
		cdc: cdc,
	}
}

// XXX: This is not the public API. This will change in MVP2 and will henceforth
// only be invoked from another module directly and not through a user
// transaction.
// TODO: Handle invalid IBC packets and return errors.
func (ibcm IBCMapper) PostIBCPacket(ctx sdk.Context,
	packet IBCPacket) sdk.Error {

	store := ctx.KVStore(ibcm.key)
	bz := ibcm.cdc.MarshalBinaryPanic(packet)

	index := ibcm.getEgressLength(store, packet.DestChain)
	store.Set(EgressKey(packet.DestChain, index), bz)
	bz = ibcm.cdc.MarshalBinary(int64(index + 1))
	store.Set(EgressLengthKey(packet.DestChain), bz)

	return nil
}

// XXX: In the future every module is able to register it's own handler for
// handling it's own IBC packets. The "ibc" handler will only route the packets
// to the appropriate callbacks.
// XXX: For now this handles all interactions with the CoinKeeper.
// XXX: This needs to do some authentication checking.
func (ibcm IBCMapper) ReceiveIBCPacket(ctx sdk.Context,
	packet IBCPacket) sdk.Error {

	return nil
}

func (ibcm IBCMapper) GetIngressSequence(ctx sdk.Context,
	srcChain string) int64 {

	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(srcChain)

	bz := store.Get(key)
	if bz == nil {
		zero := ibcm.cdc.MarshalBinaryPanic(int64(0))
		store.Set(key, zero)
		return 0
	}

	var res int64
	ibcm.cdc.UnmarshalBinaryPanic(bz, &res)
	return res
}

func (ibcm IBCMapper) SetIngressSequence(ctx sdk.Context, srcChain string,
	sequence int64) {

	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(srcChain)

	bz := ibcm.cdc.MarshalBinaryPanic(sequence)
	store.Set(key, bz)
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm IBCMapper) getEgressLength(store sdk.KVStore,
	destChain string) int64 {

	bz := store.Get(EgressLengthKey(destChain))
	if bz == nil {
		zero := ibcm.cdc.MarshalBinaryPanic(int64(0))
		store.Set(EgressLengthKey(destChain), zero)
		return 0
	}

	var res int64
	ibcm.cdc.UnmarshalBinaryPanic(bz, &res)

	return res
}

// ----------------------------
// Store keys for the IBC module.

// Stores an outgoing IBC packet under "egress/chain_id/index".
func EgressKey(destChain string, index int64) []byte {
	return []byte(fmt.Sprintf("egress/%s/%d", destChain, index))
}

// Stores the number of outgoing IBC packets under "egress/index".
func EgressLengthKey(destChain string) []byte {
	return []byte(fmt.Sprintf("egress/%s", destChain))
}

// Stores the sequence number of incoming IBC packet under "ingress/index".
func IngressSequenceKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("ingress/%s", srcChain))
}
