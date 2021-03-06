// contract
package bumo

import (
	"encoding/json"
	"errors"

	"github.com/bumoproject/bumo-sdk-go/src/3rd/proto"
	"github.com/bumoproject/bumo-sdk-go/src/keypair"
	"github.com/bumoproject/bumo-sdk-go/src/protocol"
)

type ContractOperation struct {
	url string
}

//创建合约账户
func (contract *ContractOperation) Create(sourceAddress string, destAddress string, initBalance int64, payload string, input string) ([]byte, Error) {
	if initBalance < 0 {
		return nil, sdkErr(INVALID_INITBALANCE)
	}
	if !keypair.CheckAddress(destAddress) {
		return nil, sdkErr(INVALID_DESTADDRESS)
	}
	if sourceAddress != "" {
		if !keypair.CheckAddress(sourceAddress) {
			return nil, sdkErr(INVALID_SOURCEADDRESS)
		}
	}
	if sourceAddress == destAddress {
		return nil, sdkErr(DESTADDRESS_EQUAL_SOURCEADDRESS)
	}
	if payload == "" {
		return nil, sdkErr(INVALID_PAYLOAD)
	}
	Operations := []*protocol.Operation{
		{
			SourceAddress: sourceAddress,
			Type:          protocol.Operation_CREATE_ACCOUNT,
			CreateAccount: &protocol.OperationCreateAccount{
				DestAddress: destAddress,
				Contract: &protocol.Contract{
					Payload: payload,
				},
				InitBalance: initBalance,
				InitInput:   input,
				Priv: &protocol.AccountPrivilege{
					MasterWeight: 0,
					Thresholds: &protocol.AccountThreshold{
						TxThreshold: 1,
					},
				},
			},
		},
	}
	data, err := proto.Marshal(Operations[0])
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return nil, Err
	}
	Err.Code = SUCCESS
	Err.Err = nil
	return data, Err

}

//获取合约
func (contract *ContractOperation) GetContract(address string) (string, Error) {
	if !keypair.CheckAddress(address) {
		return "", sdkErr(INVALID_PARAMETER)
	}
	get := "/getAccount?address="
	response, Err := getRequest(contract.url, get, address)
	if Err.Err != nil {
		return "", Err
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		data := make(map[string]interface{})
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err := decoder.Decode(&data)
		if err != nil {
			Err.Code = DECODER_DECODE_ERROR
			Err.Err = err
			return "", Err
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			contract := make(map[string]interface{})
			contract["contract"] = result["contract"]
			Mdata, err := json.Marshal(&contract)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			if data["error_code"].(json.Number) == "4" {
				return "", sdkErr(ACCOUNT_NOT_EXIST)
			}
			return "", getErr(data["error_code"].(float64))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err

	}
}

//转移资产并触发合约
func (contract *ContractOperation) InvokeContractByAsset(sourceAddress string, destAddress string, issueAddress string, amount int64, code string, input string) ([]byte, Error) {
	if sourceAddress != "" {
		if !keypair.CheckAddress(sourceAddress) {
			return nil, sdkErr(INVALID_SOURCEADDRESS)
		}
	}
	if !keypair.CheckAddress(destAddress) {
		return nil, sdkErr(INVALID_DESTADDRESS)
	}
	if sourceAddress == destAddress {
		return nil, sdkErr(DESTADDRESS_EQUAL_SOURCEADDRESS)
	}
	if amount < 0 {
		return nil, sdkErr(INVALID_AMOUNT)
	}
	if issueAddress != "" && !keypair.CheckAddress(issueAddress) {
		return nil, sdkErr(INVALID_ISSUEADDRESS)
	}
	if len([]rune(code)) > 64 {
		return nil, sdkErr(INVALID_CODE)
	}
	var PayAsset protocol.OperationPayAsset
	if code != "" && issueAddress != "" && amount > 0 {
		PayAsset = protocol.OperationPayAsset{
			DestAddress: destAddress,
			Asset: &protocol.Asset{
				Key: &protocol.AssetKey{
					Issuer: issueAddress,
					Code:   code,
				},
				Amount: amount,
			},
			Input: input,
		}
	} else {
		PayAsset = protocol.OperationPayAsset{
			DestAddress: destAddress,
			Input:       input,
		}
	}

	Operations := []*protocol.Operation{
		{
			SourceAddress: sourceAddress,
			Type:          protocol.Operation_PAY_ASSET,
			PayAsset:      &PayAsset,
		},
	}
	data, err := proto.Marshal(Operations[0])
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return nil, Err
	}
	Err.Code = SUCCESS
	Err.Err = nil
	return data, Err
}

//发送BU并触发合约
func (contract *ContractOperation) InvokeContractByBU(sourceAddress string, destAddress string, amount int64, input string) ([]byte, Error) {
	if sourceAddress != "" {
		if !keypair.CheckAddress(sourceAddress) {
			return nil, sdkErr(INVALID_SOURCEADDRESS)
		}
	}
	if !keypair.CheckAddress(destAddress) {
		return nil, sdkErr(INVALID_DESTADDRESS)
	}
	if sourceAddress == destAddress {
		return nil, sdkErr(DESTADDRESS_EQUAL_SOURCEADDRESS)
	}
	if amount < 0 {
		return nil, sdkErr(INVALID_AMOUNT)
	}

	Operations := []*protocol.Operation{
		{
			SourceAddress: sourceAddress,
			Type:          protocol.Operation_PAY_COIN,
			PayCoin: &protocol.OperationPayCoin{
				DestAddress: destAddress,
				Amount:      amount,
				Input:       input,
			},
		},
	}
	data, err := proto.Marshal(Operations[0])
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return nil, Err
	}
	Err.Code = SUCCESS
	Err.Err = nil
	return data, Err
}
