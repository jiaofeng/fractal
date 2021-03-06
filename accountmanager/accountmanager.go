// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package accountmanager

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var acctInfoPrefix = "AcctInfo"

// AccountManager represents account management model.
type AccountManager struct {
	sdb SdbIf
	ast *asset.Asset
}

//NewAccountManager create new account manager
func NewAccountManager(db *state.StateDB) (*AccountManager, error) {
	if db == nil {
		return nil, ErrNewAccountErr
	}
	return &AccountManager{
		sdb: db,
		ast: asset.NewAsset(db),
	}, nil
}

// AccountIsExist check account is exist.
func (am *AccountManager) AccountIsExist(accountName common.Name) (bool, error) {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct != nil {
		return true, nil
	}
	return false, nil
}

//AccountIsEmpty check code size > 0
func (am *AccountManager) AccountIsEmpty(accountName common.Name) (bool, error) {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct == nil {
		return false, ErrAccountNotExist
	}

	if acct.IsEmpty() {
		return true, nil
	}
	return false, nil
}

//CreateAccount contract account pubkey = nil
func (am *AccountManager) CreateAccount(accountName common.Name, pubkey common.PubKey) error {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct != nil {
		return ErrAccountIsExist
	}

	acctObj, err := NewAccount(accountName, pubkey)
	if err != nil {
		return err
	}
	if acctObj == nil {
		return ErrCreateAccountError
	}

	am.SetAccount(acctObj)
	return nil
}

//UpdateAccount update the pubkey of the accunt
func (am *AccountManager) UpdateAccount(accountName common.Name, pubkey common.PubKey) error {
	acct, err := am.GetAccountByName(accountName)
	if acct == nil {
		return ErrAccountNotExist
	}
	if err != nil {
		return err
	}
	acct.SetPubKey(pubkey)
	return am.SetAccount(acct)
}

//GetAccountByName get account by name
func (am *AccountManager) GetAccountByName(accountName common.Name) (*Account, error) {
	b, err := am.sdb.Get(accountName.String(), acctInfoPrefix)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}

	var acct Account
	if err := rlp.DecodeBytes(b, &acct); err != nil {
		return nil, err
	}

	//user can find destroyed account
	//if acct.IsDestoryed() == true {
	//	return nil, ErrAccountNotExist
	//}

	return &acct, nil
}

//store account object to db
func (am *AccountManager) SetAccount(acct *Account) error {
	if acct == nil {
		return ErrAccountIsNil
	}
	if acct.IsDestoryed() == true {
		return ErrAccountIsDestroy
	}
	b, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return err
	}
	am.sdb.Put(acct.GetName().String(), acctInfoPrefix, b)
	return nil
}

//DeleteAccountByName delete account
func (am *AccountManager) DeleteAccountByName(accountName common.Name) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return ErrAccountNotExist
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	acct.SetDestory()
	b, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return err
	}
	am.sdb.Put(acct.GetName().String(), acctInfoPrefix, b)
	return nil
}

// GetNonce get nonce
func (am *AccountManager) GetNonce(accountName common.Name) (uint64, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	return acct.GetNonce(), nil
}

// SetNonce set nonce
func (am *AccountManager) SetNonce(accountName common.Name, nonce uint64) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	acct.SetNonce(nonce)
	return am.SetAccount(acct)
}

//GetBalancesList get Balances return a list
//func (am *AccountManager) GetBalancesList(accountName common.Name) ([]*AssetBalance, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	return acct.GetBalancesList(), nil
//}

//GetAllAccountBalance return all balance in map.
//func (am *AccountManager) GetAccountAllBalance(accountName common.Name) (map[uint64]*big.Int, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	if acct == nil {
//		return nil, ErrAccountNotExist
//	}
//
//	return acct.GetAllBalances()
//}

//GetAcccountPubkey get account pub key
//func (am *AccountManager) GetAcccountPubkey(accountName common.Name) ([]byte, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	if acct == nil {
//		return nil, ErrAccountNotExist
//	}
//	return acct.GetPubKey().Bytes(), nil
//}

// RecoverTx Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer types.Signer, tx *types.Transaction) error {
	for _, action := range tx.GetActions() {
		pub, err := types.Recover(signer, action, tx)
		if err != nil {
			return err
		}

		if err := am.IsValidSign(action.Sender(), action.Type(), pub); err != nil {
			return err
		}
	}
	return nil
}

// IsValidSign
func (am *AccountManager) IsValidSign(accountName common.Name, aType types.ActionType, pub common.PubKey) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	if acct.IsDestoryed() {
		return ErrAccountIsDestroy
	}
	//TODO action type verify

	if acct.GetPubKey().Compare(pub) != 0 {
		return fmt.Errorf("%v %v have %v excepted %v", acct.AcctName, ErrkeyNotSame, acct.GetPubKey().String(), pub.String())
	}
	return nil

}

//GetAssetInfoByName get asset info by asset name.
func (am *AccountManager) GetAssetInfoByName(assetName string) (*asset.AssetObject, error) {
	assetID, err := am.ast.GetAssetIdByName(assetName)
	if err != nil {
		return nil, err
	}
	return am.ast.GetAssetObjectById(assetID)
}

//GetAssetInfoByID get asset info by assetID
func (am *AccountManager) GetAssetInfoByID(assetID uint64) (*asset.AssetObject, error) {
	return am.ast.GetAssetObjectById(assetID)
}

//GetAccountBalanceByID get account balance by ID
func (am *AccountManager) GetAccountBalanceByID(accountName common.Name, assetID uint64) (*big.Int, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return big.NewInt(0), err
	}
	if acct == nil {
		return big.NewInt(0), ErrAccountNotExist
	}
	return acct.GetBalanceByID(assetID)
}

//GetAccountBalanceByName get account balance by name
//func (am *AccountManager) GetAccountBalanceByName(accountName common.Name, assetName string) (*big.Int, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//	if acct == nil {
//		return big.NewInt(0), ErrAccountNotExist
//	}
//
//	assetID, err := am.ast.GetAssetIdByName(assetName)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//	if assetID == 0 {
//		return big.NewInt(0), asset.ErrAssetNotExist
//	}
//
//	ba := &big.Int{}
//	ba, err = acct.GetBalanceByID(assetID)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//
//	return ba, nil
//}

//SubAccountBalanceByID sub balance by assetID
func (am *AccountManager) SubAccountBalanceByID(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	//
	val, err := acct.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	acct.SetBalance(assetID, new(big.Int).Sub(val, value))
	return am.SetAccount(acct)
}

//AddAccountBalanceByID add balance by assetID
func (am *AccountManager) AddAccountBalanceByID(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := acct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		acct.AddNewAssetByAssetID(assetID, value)
	} else {
		acct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	return am.SetAccount(acct)
}

func (am *AccountManager) AddAccountBalanceByName(accountName common.Name, assetName string, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	assetID, err := am.ast.GetAssetIdByName(assetName)
	if err != nil {
		return err
	}

	if assetID == 0 {
		return asset.ErrAssetNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := acct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		acct.AddNewAssetByAssetID(assetID, value)
	} else {
		acct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	return am.SetAccount(acct)
}

//
func (am *AccountManager) EnoughAccountBalance(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	return acct.EnoughAccountBalance(assetID, value)
}

//
func (am *AccountManager) GetCode(accountName common.Name) ([]byte, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, ErrAccountNotExist
	}
	return acct.GetCode()
}

////
//func (am *AccountManager) SetCode(accountName common.Name, code []byte) (bool, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return false, err
//	}
//	if acct == nil {
//		return false, ErrAccountNotExist
//	}
//	err = acct.SetCode(code)
//	if err != nil {
//		return false, err
//	}
//	err = am.SetAccount(acct)
//	if err != nil {
//		return false, err
//	}
//	return true, nil
//}

//
//GetCodeSize get code size
func (am *AccountManager) GetCodeSize(accountName common.Name) (uint64, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	return acct.GetCodeSize(), nil
}

// GetCodeHash get code hash
//func (am *AccountManager) GetCodeHash(accountName common.Name) (common.Hash, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	if acct == nil {
//		return common.Hash{}, ErrAccountNotExist
//	}
//	return acct.GetCodeHash()
//}

//GetAccountFromValue  get account info via value bytes
func (am *AccountManager) GetAccountFromValue(accountName common.Name, key string, value []byte) (*Account, error) {
	if len(value) == 0 {
		return nil, ErrAccountNotExist
	}
	if key != accountName.String()+acctInfoPrefix {
		return nil, ErrAccountNameInvalid
	}
	var acct Account
	if err := rlp.DecodeBytes(value, &acct); err != nil {
		return nil, ErrAccountNotExist
	}
	if !common.IsSameName(acct.AcctName, accountName) {
		return nil, ErrAccountNameInvalid
	}
	return &acct, nil
}

// CanTransfer check if can transfer.
func (am *AccountManager) CanTransfer(accountName common.Name, assetID uint64, value *big.Int) (bool, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if err = acct.EnoughAccountBalance(assetID, value); err == nil {
		return true, nil
	}
	return false, err
}

//TransferAsset
func (am *AccountManager) TransferAsset(fromAccount common.Name, toAccount common.Name, assetID uint64, value *big.Int) error {
	fromAcct, err := am.GetAccountByName(fromAccount)
	if err != nil {
		return err
	}
	if fromAcct == nil {
		return ErrAccountNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	if common.IsSameName(fromAccount, toAccount) {
		return nil
	}
	val, err := fromAcct.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	fromAcct.SetBalance(assetID, new(big.Int).Sub(val, value))

	toAcct, err := am.GetAccountByName(toAccount)
	if err != nil {
		return err
	}
	if toAcct == nil {
		return ErrAccountNotExist
	}
	if toAcct.IsDestoryed() {
		return ErrAccountIsDestroy
	}
	val, err = toAcct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		toAcct.AddNewAssetByAssetID(assetID, value)
	} else {
		toAcct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	if err = am.SetAccount(fromAcct); err != nil {
		return err
	}
	return am.SetAccount(toAcct)
}

//IssueAsset issue asset
func (am *AccountManager) IssueAsset(asset *asset.AssetObject) error {
	if err := am.ast.IssueAsset(asset.GetAssetName(), asset.GetSymbol(), asset.GetAssetAmount(), asset.GetDecimals(), asset.GetAssetOwner()); err != nil {
		return err
	}
	acct, err := am.GetAccountByName(asset.GetAssetOwner())
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	return am.AddAccountBalanceByName(asset.GetAssetOwner(), asset.GetAssetName(), asset.GetAssetAmount())
}

//increase asset and add amount to accout balance
func (am *AccountManager) IncAsset2Acct(fromName common.Name, toName common.Name, assetID uint64, amount *big.Int) error {
	if err := am.ast.IncreaseAsset(fromName, assetID, amount); err != nil {
		return err
	}
	return am.AddAccountBalanceByID(toName, assetID, amount)
}

//AddBalanceByName add balance to account
//func (am *AccountManager) AddBalanceByName(accountName common.Name, assetID uint64, amount *big.Int) error {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return err
//	}
//	if acct == nil {
//		return ErrAccountNotExist
//	}
//	return acct.AddBalanceByID(assetID, amount)
//	rerturn
//}

// Process account action

func (am *AccountManager) Process(action *types.Action) error {
	snap := am.sdb.Snapshot()
	err := am.process(action)
	if err != nil {
		am.sdb.RevertToSnapshot(snap)
	}
	return err
}

func (am *AccountManager) process(action *types.Action) error {
	switch action.Type() {
	case types.CreateAccount:
		var key common.PubKey
		key.SetBytes(action.Data())
		if err := am.CreateAccount(action.Recipient(), key); err != nil {
			return err
		}
		break
	case types.UpdateAccount:
		var key common.PubKey
		key.SetBytes(action.Data())
		if err := am.UpdateAccount(action.Sender(), key); err != nil {
			return err
		}
		break
	//case types.DeleteAccount:
	//	if err := am.DeleteAccountByName(action.Sender()); err != nil {
	//		return err
	//	}
	//	break
	case types.IssueAsset:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		if err := am.IssueAsset(&asset); err != nil {
			return err
		}
		break
	case types.IncreaseAsset:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		if err = am.IncAsset2Acct(action.Sender(), action.Sender(), asset.GetAssetId(), asset.GetAssetAmount()); err != nil {
			return err
		}
		break
	case types.SetAssetOwner:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		acct, err := am.GetAccountByName(asset.GetAssetOwner())
		if err != nil {
			return err
		}
		if acct == nil {
			return ErrAccountNotExist
		}
		if err := am.ast.SetAssetNewOwner(action.Sender(), asset.GetAssetId(), asset.GetAssetOwner()); err != nil {
			return err
		}
		break
	case types.Transfer:
		return am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value())
	default:
		return ErrUnkownTxType
	}

	if action.Value().Cmp(big.NewInt(0)) > 0 {
		return am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value())
	}
	return nil
}
