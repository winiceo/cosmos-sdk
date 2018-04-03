package types

import (
	"bytes"
	"fmt"
	"reflect"

	oldwire "github.com/tendermint/go-wire"

	wire "github.com/cosmos/cosmos-sdk/wire"
)

// This AccountMapper encodes/decodes accounts from stores retrieved from
// the context using the go-wire (binary) encoding/decoding library.
type AccountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key StoreKey

	// The prototypical Account concrete type.
	proto Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new AccountMapper that
// uses go-wire to (binary) encode and decode concrete Accounts.
func NewAccountMapper(key StoreKey, proto Account, additional ...interface{}) AccountMapper {
	cdc := wire.NewCodec()

	RegisterWireProtoAccount(cdc)

	for _, obj := range additional {
		cdc.RegisterInterface((*MyInterface2)(nil), nil)
		cdc.RegisterConcrete(SendMsg{}, "github.com/cosmos/cosmos-sdk/bank/SendMsg", nil)
	}

	//cdc.RegisterConcrete(SendMsg{}, "github.com/cosmos/cosmos-sdk/bank/SendMsg", nil)

	return AccountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

func (am AccountMapper) NewAccountWithAddress(ctx Context, addr Address) Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

func (am AccountMapper) GetAccount(ctx Context, addr Address) Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

func (am AccountMapper) SetAccount(ctx Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(am.key)
	bz := am.encodeAccount(acc)
	store.Set(addr, bz)
}

//----------------------------------------
// misc.

// NOTE: currently unused
func (am AccountMapper) clonePrototypePtr() interface{} {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoErt := protoRt.Elem()
		if protoErt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto Account, or a pointer to one")
		}
		protoRv := reflect.New(protoErt)
		return protoRv.Interface()
	} else {
		protoRv := reflect.New(protoRt)
		return protoRv.Interface()
	}
}

// Creates a new struct (or pointer to struct) from am.proto.
func (am AccountMapper) clonePrototype() Account {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto Account, but %v doesn't implement Account", protoRt))
		}
		return clone
	} else {
		protoRv := reflect.New(protoRt).Elem()
		clone, ok := protoRv.Interface().(Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto Account, but %v doesn't implement Account", protoRt))
		}
		return clone
	}
}

func (am AccountMapper) encodeAccount(acc Account) []byte {
	bz, err := am.cdc.MarshalBinary(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AccountMapper) decodeAccount(bz []byte) Account {
	// ... old go-wire ...
	r, n, err := bytes.NewBuffer(bz), new(int), new(error)
	accI := oldwire.ReadBinary(struct{ Account }{}, r, len(bz), n, err)
	if *err != nil {
		panic(*err)
	}

	acc := accI.(struct{ Account }).Account
	return acc

	/*
		accPtr := am.clonePrototypePtr()
			err := am.cdc.UnmarshalBinary(bz, accPtr)
			if err != nil {
				panic(err)
			}
			if reflect.ValueOf(am.proto).Kind() == reflect.Ptr {
				return reflect.ValueOf(accPtr).Interface().(Account)
			} else {
				return reflect.ValueOf(accPtr).Elem().Interface().(Account)
			}
	*/
}

//----------------------------------------
// Wire

func RegisterWireProtoAccount(cdc *wire.Codec) {
	// Register crypto.[PubKey,PrivKey,Signature] types.
	wire.RegisterCrypto(cdc)
}
