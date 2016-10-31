/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var A, B string    // Entities
	var Aval int // Asset holdings
	var err error

	if len(args) % 2 != 1 {
		return nil, errors.New("Expecting odd number of arguments: [name, bank1, amount1, bank2, amount2, ......]")
	}

	if len(args) <= 1 {
		return nil, errors.New("Too few arguments. Expecting at least 3 (at least one bank account)")
	}

	// Initialize the chaincode
	A = args[0]
	bankList := ""
	for i := 1; i < len(args); i += 2 {
		B = args[i]
		Aval, err = strconv.Atoi(args[i + 1])
		if err != nil {
			return nil, errors.New("Expecting integer value for asset holding")
		}
		fmt.Printf("Bank = %s, Aval = %d\n", B, Aval)
	
		// Write the state to the ledger
		key := A + "_" + B
		err = stub.PutState(key, []byte(strconv.Itoa(Aval)))
		if err != nil {
			return nil, err
		}

		if bankList != "" {
			bankList += ","
		}
		bankList += B
	}
	// Write the state to the ledger
	err = stub.PutState(A, []byte(bankList))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var A, B, method string    // Entities
	var Aval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	method = args[0]
	A = args[1]
	B = args[3]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	key := A + "_" + B
	Avalbytes, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))


	// Perform the execution
	X, err = strconv.Atoi(args[2])
	
	if method == "deposit" {
		Aval = Aval + X
	}else if method == "withdraw"{

		if Aval - X > 0 {
			Aval = Aval - X
		}else{
			Aval = 0
		}
	}

	fmt.Printf("Aval = %d\n", Aval)

	// Write the state back to the ledger
	err = stub.PutState(key, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Run callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	// Handle different functions
	if function == "init" {
		// Initialize the entities and their asset holdings
		return t.Init(stub, function, args)
	} else if function == "invoke" {
		// Transaction makes payment of X units from A to B
		return t.Invoke(stub, function, args)
	}

	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var A, B string // Entities
	var err error

	if len(args) > 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the account holder followed by the bank name")
	}

	A = args[0]
	key := A
	keyName := "List"
	if len(args) == 2 {
		B = args[1]
		key = A + "_" + B
		keyName = "Amount"
	}

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + key + "\",\"" + keyName + "\":\"" + string(Avalbytes) + "\"}"

	fmt.Printf("Query Response:%s\n", jsonResp)
	return []byte(jsonResp), nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
