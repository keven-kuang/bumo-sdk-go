// bumo
package bumo

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/bumoproject/bumo-sdk-go/src/3rd/proto"
	"github.com/bumoproject/bumo-sdk-go/src/keypair"
	"github.com/bumoproject/bumo-sdk-go/src/protocol"
	"github.com/bumoproject/bumo-sdk-go/src/signature"
)

const Conversion float64 = 100000000

type deal struct {
	Items []Items `json:"items"`
}
type Items struct {
	Transaction_blob string       `json:"transaction_blob"`
	Signatures       []Signatures `json:"signatures"`
}
type Signatures struct {
	Sign_data  string `json:"sign_data"`
	Public_key string `json:"public_key"`
}

type BumoSdk struct {
	Account  AccountOperation
	Contract ContractOperation
}

//新建链接
func (bumosdk *BumoSdk) Newbumo(ip string) Error {
	if ip == "" {
		return sdkErr(INVALID_PARAMETER)
	}
	bumosdk.Account.url = ip
	bumosdk.Contract.url = ip
	Err.Code = SUCCESS
	Err.Err = nil
	return Err
}

//获取区块高度
func (bumosdk *BumoSdk) GetBlockNumber() (int64, Error) {
	get := "/getLedger"
	response, Err := getRequest(bumosdk.Account.url, get, "")
	if Err.Err != nil {
		return 0, Err
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
			return 0, Err
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			header := result["header"].(map[string]interface{})
			seqstr := header["seq"].(json.Number)
			seq, err := strconv.ParseInt(string(seqstr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return seq, Err
		} else {
			if data["error_code"].(json.Number) == "4" {
				return 0, sdkErr(BLOCK_NOT_EXIST)
			}
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, Err
			}
			return 0, getErr(float64(errorCode))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return 0, Err
	}
}

//检查区块同步
func (bumosdk *BumoSdk) CheckBlockStatus() (bool, Error) {
	var ret bool
	get := "getModulesStatus"
	response, Err := getRequest(bumosdk.Account.url, get, "")
	if Err.Err != nil {
		return false, Err
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
			return false, Err
		}
		ledger_manager := data["ledger_manager"].(map[string]interface{})
		if ledger_manager["chain_max_ledger_seq"] == ledger_manager["ledger_sequence"] {
			ret = true
		}
	}
	Err.Code = SUCCESS
	Err.Err = nil
	return ret, Err
}

//根据hash查询交易
func (bumosdk *BumoSdk) GetTransaction(transactionHash string) (string, Error) {

	if len(transactionHash) != 64 {
		return "", sdkErr(INVALID_PARAMETER)
	}
	get := "/getTransactionHistory?hash="
	response, Err := getRequest(bumosdk.Account.url, get, transactionHash)
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
			result := data["result"]
			Mdata, err := json.Marshal(&result)
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
				return "", sdkErr(TRANSACTION_NOT_EXIST)
			}
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			return "", getErr(float64(errorCode))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//根据高度查询交易
func (bumosdk *BumoSdk) GetBlock(blockNumber int64) (string, Error) {
	if blockNumber < 0 {
		return "", sdkErr(INVALID_PARAMETER)
	}
	bnstr := strconv.FormatInt(blockNumber, 10)
	get := "/getTransactionHistory?ledger_seq="
	response, Err := getRequest(bumosdk.Account.url, get, bnstr)
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
			result := data["result"]
			Mdata, err := json.Marshal(&result)
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
				return "", sdkErr(BLOCK_NOT_EXIST)
			}
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			return "", getErr(float64(errorCode))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//查询区块头
func (bumosdk *BumoSdk) GetLedger(blockNumber int64) (string, Error) {
	if blockNumber <= 0 {
		return "", sdkErr(INVALID_PARAMETER)
	}
	bnstr := strconv.FormatInt(blockNumber, 10)
	get := "/getLedger?seq="
	response, Err := getRequest(bumosdk.Account.url, get, bnstr)
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
			result := data["result"]
			Mdata, err := json.Marshal(&result)
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
				return "", sdkErr(BLOCK_NOT_EXIST)
			}
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			return "", getErr(float64(errorCode))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//生成交易(默认费用)
func (bumosdk *BumoSdk) createTransactionWithDefaultFee(sourceAddress string, nonce int64, operation []byte) (string, Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return "", sdkErr(INVALID_SOURCEADDRESS)
	}
	if nonce <= 0 {
		return "", sdkErr(INVALID_NONCE)
	}
	if operation == nil {
		return "", sdkErr(INVALID_OPERATION)
	}
	var feeLimit int64
	gasPrice, _, Err := getFees(bumosdk.Account.url)
	if Err.Err != nil {
		return "", Err
	}
	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	if operations.Type == protocol.Operation_ISSUE_ASSET {
		feeLimit = (5000000 + 1000) * gasPrice
	} else if operations.Type == protocol.Operation_CREATE_ACCOUNT {
		feeLimit = (1000000 + 1000) * gasPrice
	} else {
		feeLimit = 1000 * gasPrice
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	Transaction := &protocol.Transaction{
		SourceAddress: sourceAddress,
		Nonce:         nonce,
		FeeLimit:      feeLimit,
		GasPrice:      gasPrice,
		Operations:    Operations,
	}
	data, err := proto.Marshal(Transaction)
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	dataHash := hex.EncodeToString(data)
	Err.Code = SUCCESS
	Err.Err = nil
	return dataHash, Err
}

//生成交易(传入费用)
func (bumosdk *BumoSdk) CreateTransactionWithFee(sourceAddress string, nonce int64, gasPrice int64, feeLimit int64, operation []byte) (string, Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return "", sdkErr(INVALID_SOURCEADDRESS)
	}
	newgasPrice, _, Err := getFees(bumosdk.Account.url)
	if Err.Err != nil {
		return "", Err
	}
	if nonce <= 0 {
		return "", sdkErr(INVALID_NONCE)
	}
	if gasPrice < newgasPrice {
		return "", sdkErr(INVALID_GASPRICE)
	}
	if feeLimit < newgasPrice*1000 {
		return "", sdkErr(INVALID_FEELIMIT)
	}
	if operation == nil {
		return "", sdkErr(INVALID_OPERATION)
	}
	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	Transaction := &protocol.Transaction{
		SourceAddress: sourceAddress,
		Nonce:         nonce,
		FeeLimit:      feeLimit,
		GasPrice:      gasPrice,
		Operations:    Operations,
	}
	data, err := proto.Marshal(Transaction)
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	dataHash := hex.EncodeToString(data)
	Err.Code = SUCCESS
	Err.Err = nil
	return dataHash, Err
}

//评估费用
func (bumosdk *BumoSdk) EvaluationFee(sourceAddress string, nonce int64, operation []byte, signatureNumber int64) (int64, int64, Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return 0, 0, sdkErr(INVALID_SOURCEADDRESS)
	}
	if nonce <= 0 {
		return 0, 0, sdkErr(INVALID_NONCE)
	}
	if operation == nil {
		return 0, 0, sdkErr(INVALID_OPERATION)
	}
	if signatureNumber <= 0 {
		return 0, 0, sdkErr(INVALID_SIGNATURENUMBER)
	}

	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return 0, 0, Err
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	request := make(map[string]interface{})
	transactionJson := make(map[string]interface{})
	transactionJson["source_address"] = sourceAddress
	transactionJson["nonce"] = nonce
	transactionJson["operations"] = Operations
	transactionJson["signature_number"] = signatureNumber
	items := make([]map[string]interface{}, 1)
	items[0] = make(map[string]interface{})
	items[0]["transaction_json"] = transactionJson
	request["items"] = items
	requestJson, err := json.Marshal(request)
	if err != nil {
		Err.Code = JSON_MARSHAL_ERROR
		Err.Err = err
		return 0, 0, Err
	}
	post := "/testTransaction"
	response, Err := postRequest(bumosdk.Account.url, post, requestJson)
	if Err.Err != nil {
		return 0, 0, Err
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		data := make(map[string]interface{})
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err = decoder.Decode(&data)
		if err != nil {
			Err.Code = DECODER_DECODE_ERROR
			Err.Err = err
			return 0, 0, Err
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			txs, ok := result["txs"].([]interface{})
			if !ok {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			tx, ok := txs[0].(map[string]interface{})
			if !ok {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			if tx["actual_fee"] == nil {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			actualFeestr := tx["actual_fee"].(json.Number)
			actualFee, err := strconv.ParseInt(string(actualFeestr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			transactionEnv := tx["transaction_env"].(map[string]interface{})
			transaction := transactionEnv["transaction"].(map[string]interface{})
			if transaction["gas_price"] == nil {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			gasPriceStr := transaction["gas_price"].(json.Number)
			gasPrice, err := strconv.ParseInt(string(gasPriceStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return int64(actualFee), int64(gasPrice), Err
		} else {
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			Err.Code = int(float64(errorCode) + 10000)
			Err.Err = errors.New(data["error_desc"].(string))
			return 0, 0, Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return 0, 0, Err
	}
}

//签名
func (bumosdk *BumoSdk) SignTransaction(transactionBlob string, privateKey string) (string, string, Error) {
	if transactionBlob == "" {
		return "", "", sdkErr(INVALID_TRANSACTIONBLOB)
	}
	if !keypair.CheckPrivateKey(privateKey) {
		return "", "", sdkErr(INVALID_PRIVATEKEY)
	}
	publicKey, err := keypair.GetEncPublicKey(privateKey)
	if err != nil {
		Err.Code = KEYPAIR_GETENCPUBLICKEY_ERROR
		Err.Err = err
		return "", "", Err
	}
	TransactionBlob, err := hex.DecodeString(transactionBlob)
	if err != nil {
		Err.Code = HEX_DECODESTRING_ERROR
		Err.Err = err
		return "", "", Err
	}
	signData, err := signature.Sign(privateKey, TransactionBlob)
	if err != nil {
		Err.Code = SIGNATURE_SIGN_ERROR
		Err.Err = err
		return "", "", Err
	}
	return signData, publicKey, Err
}

//多签名
func (bumosdk *BumoSdk) MultiSignTransaction(transactionBlob string, privateKey []string) ([]Signatures, Error) {
	if transactionBlob == "" {
		return nil, sdkErr(INVALID_TRANSACTIONBLOB)
	}
	for i := range privateKey {
		if !keypair.CheckPrivateKey(privateKey[i]) {
			return nil, sdkErr(INVALID_PRIVATEKEY)
		}
	}
	signatures := make([]Signatures, len(privateKey))
	var err error
	for i := range privateKey {
		signatures[i].Public_key, err = keypair.GetEncPublicKey(privateKey[i])
		if err != nil {
			Err.Code = KEYPAIR_GETENCPUBLICKEY_ERROR
			Err.Err = err
			return nil, Err
		}
	}

	TransactionBlob, err := hex.DecodeString(transactionBlob)
	if err != nil {
		Err.Code = HEX_DECODESTRING_ERROR
		Err.Err = err
		return nil, Err
	}
	for i := range privateKey {
		signatures[i].Sign_data, err = signature.Sign(privateKey[i], TransactionBlob)
		if err != nil {
			Err.Code = SIGNATURE_SIGN_ERROR
			Err.Err = err
			return nil, Err
		}
	}

	return signatures, Err
}

//单签名交易提交
func (bumosdk *BumoSdk) SubmitTransaction(transactionBlob string, signData string, publicKey string) (string, Error) {
	if signData == "" {
		return "", sdkErr(INVALID_SIGNDATA)
	}
	if transactionBlob == "" {
		return "", sdkErr(INVALID_TRANSACTIONBLOB)
	}
	if !keypair.CheckPublicKey(publicKey) {
		return "", sdkErr(INVALID_PUBLICKEY)
	}
	signatures := make([]Signatures, 1)
	signatures[0].Sign_data = signData
	signatures[0].Public_key = publicKey
	requestJson, Err := getRequestJson(transactionBlob, signatures)
	if Err.Err != nil {
		return "", Err
	}
	post := "/submitTransaction"
	response, Err := postRequest(bumosdk.Account.url, post, requestJson)
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
		results := data["results"].([]interface{})
		result := results[0].(map[string]interface{})
		if result["error_code"].(json.Number) == "0" {
			hash := make(map[string]interface{})
			hash["hash"] = result["hash"]
			Mdata, err := json.Marshal(&hash)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			errorCodeStr := result["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = int(float64(errorCode) + 10000)
			Err.Err = errors.New(result["error_desc"].(string))
			return "", Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//多签名交易提交
func (bumosdk *BumoSdk) SubmitTransWithMultiSign(transactionBlob string, signatures []Signatures) (string, Error) {
	if transactionBlob == "" {
		return "", sdkErr(INVALID_TRANSACTIONBLOB)
	}
	for i := range signatures {
		if !keypair.CheckPublicKey(signatures[i].Public_key) || signatures[i].Sign_data == "" {
			return "", sdkErr(INVALID_SIGNATURES)
		}
	}
	requestJson, Err := getRequestJson(transactionBlob, signatures)
	if Err.Err != nil {
		return "", Err
	}
	post := "/submitTransaction"
	response, Err := postRequest(bumosdk.Account.url, post, requestJson)
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
		results := data["results"].([]interface{})
		result := results[0].(map[string]interface{})
		if result["error_code"].(json.Number) == "0" {
			hash := make(map[string]interface{})
			hash["hash"] = result["hash"]
			Mdata, err := json.Marshal(&hash)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			errorCodeStr := result["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = int(float64(errorCode) + 10000)
			Err.Err = errors.New(result["error_desc"].(string))
			return "", Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//获取最新fees
func getFees(url string) (int64, int64, Error) {
	get := "/getLedger?with_fee=true"
	response, Err := getRequest(url, get, "")
	if Err.Err != nil {
		return 0, 0, Err
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
			return 0, 0, Err
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			fees := result["fees"].(map[string]interface{})
			gasPriceStr, ok := fees["gas_price"].(json.Number)
			if ok != true {
				Err.Code = SUCCESS
				Err.Err = nil
				return 0, 0, Err
			}
			baseReserveStr, ok := fees["base_reserve"].(json.Number)
			if ok != true {
				Err.Code = SUCCESS
				Err.Err = nil
				return 0, 0, Err
			}
			gasPrice, err := strconv.ParseInt(string(gasPriceStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			baseReserve, err := strconv.ParseInt(string(baseReserveStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return gasPrice, baseReserve, Err
		} else {
			errorCodeStr := data["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodeStr), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return 0, 0, Err
			}
			Err.Code = int(float64(errorCode) + 10000)
			Err.Err = errors.New(data["error_desc"].(string))
			return 0, 0, Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return 0, 0, Err
	}
}
