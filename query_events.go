/*
 * Flow Go SDK
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"github.com/onflow/flow-go-sdk/templates"

	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/examples"
)

func main() {
	QueryEventsDemo()
}

func QueryEventsDemo() {
	ctx := context.Background()

	flowClient, err := client.New("127.0.0.1:3569", grpc.WithInsecure())
	examples.Handle(err)

	acctAddr, acctKey, acctSigner := examples.RandomAccount(flowClient)

	// 关于合约的开发和部署,之后的文章会有更详细的讲解,本文不做展开.
	contract := `
		pub contract EventDemo {
			pub event Add(x: Int, y: Int, sum: Int)

			pub fun add(x: Int, y: Int) {
				let sum = x + y
				emit Add(x: x, y: y, sum: sum)
			}
		}
	`

	contractAccount := examples.CreateAccountWithContracts(flowClient,
		nil, []templates.Contract{{
			Name:   "EventDemo",
			Source: contract,
		}})

	script := fmt.Sprintf(`
		import EventDemo from 0x%s

		transaction {
			execute {
				EventDemo.add(x: 2, y: 3)
			}
		}
	`, contractAccount.Address.Hex())

	referenceBlockID := examples.GetReferenceBlockId(flowClient)
	// 构建交易
	runScriptTx := flow.NewTransaction().
		SetScript([]byte(script)).
		SetPayer(acctAddr).
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(acctAddr, acctKey.Index, acctKey.SequenceNumber)

	err = runScriptTx.SignEnvelope(acctAddr, acctKey.Index, acctSigner)
	examples.Handle(err)

	// 发送交易
	err = flowClient.SendTransaction(ctx, *runScriptTx)
	examples.Handle(err)

	examples.WaitForSeal(ctx, flowClient, runScriptTx.ID())

	result, err := flowClient.GetTransactionResult(ctx, runScriptTx.ID())
	examples.Handle(err)

	fmt.Println("\nQuery for tx by hash:")
	for i, event := range result.Events {
		fmt.Printf("Found event #%d\n", i+1)
		fmt.Printf("Transaction ID: %s\n", event.TransactionID)
		fmt.Printf("Event ID: %s\n", event.ID())
		fmt.Println(event.String())
	}
}
