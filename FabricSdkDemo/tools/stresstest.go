package tools

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	channelID      = "ambrchannel"
	orgName        = "org1"
	orgAdmin       = "Admin"
	ordererOrgName = "orderer.ambr.com"
	ccID           = "mycc"
)

// ExampleCC query and transaction arguments

func getQueryArgs(i int) [][]byte {
	str1 := strconv.Itoa(i)
	return [][]byte{[]byte(str1)}
}

func getTxArgs(i int) [][]byte {
	str1 := strconv.Itoa(i)
	return [][]byte{[]byte(str1), []byte("{\"name\":\"Tom\",\"age\":20}")}
}

//for stability testing
func setupAndRun2(configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {
	//Init the sdk config
	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		return
	}
	defer sdk.Close()
	// ************ setup complete ************** //

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	//clientChannelContext := sdk.NewClient(fabsdk.WithUser(orgAdmin)).Channel(channelID)

	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)

	if err != nil {
		fmt.Printf("Failed to create new channel client: %s\n", err)
		return
	}

	eventID := "([a-zA-Z]+)"
	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := client.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		fmt.Printf("Failed to register cc event: %s\n", err)
		return
	} //endif
	defer client.UnregisterChaincodeEvent(reg)

	start := time.Now()
	wait := sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wait.Add(1)
		go func(index int) {
			// Move funds

			sum := 3600 * 6
			for j := 0; j < sum; j++ {
				executeCC(client, sum*index+j)

				select {
				case ccEvent := <-notifier:
					fmt.Printf("Received CC event: %#v\n", ccEvent)
				case <-time.After(time.Second * 1):
					fmt.Printf("Did NOT receive CC event for eventId(%s)\n", eventID)
				}

				value := queryCC(client, sum*index+j)
				str := ""
				if value != nil {
					str = string(value)
				}
				fmt.Println("value is ", str)
			}
			wait.Done()
		}(i)
	}
	wait.Wait()

	t := time.Since(start)
	fmt.Println("Ok, 10k get & 10 set takes: ", t)
}

//for stress testing
func setupAndRun(configOpt core.ConfigProvider, threads, rounds int, sdkOpts ...fabsdk.Option) (string, error) {

	//Init the sdk config
	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		return "", err
	}
	defer sdk.Close()
	// ************ setup complete ************** //

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	//clientChannelContext := sdk.NewClient(fabsdk.WithUser(orgAdmin)).Channel(channelID)

	/*c, _ := ledger.New(clientChannelContext)
	tx2, _ := c.QueryTransaction(fab.TransactionID("e1475c4a41017ddd998bd34ca669b1f2ff528c3a1e6d8960323c0217e7bc6567"))
	fmt.Println("txxxxxxxxxxxxxxxxx: ", tx2)
	*/

	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)

	if err != nil {
		fmt.Printf("Failed to create new channel client: %s\n", err)
		return "", err
	}

	ch := make(chan int, rounds)
	for i := 0; i < rounds; i++ {
		ch <- i
	}
	fmt.Println("rounds into channel", threads, rounds)

	/*
		go func() {
			eventID := ".*"
			// Register chaincode event (pass in channel which receives event details when the event is complete)
			reg, notifier, err := client.RegisterChaincodeEvent(ccID, eventID)
			if err != nil {
				fmt.Printf("Failed to register cc event: %s\n", err)
				return
			} //endif
			defer client.UnregisterChaincodeEvent(reg)

			for {
				select {
				case notify, ok := <-notifier:
					if !ok {
						break
					}

					fmt.Println("******************: ", notify)
				}
			}
		}()
	*/

	start := time.Now()
	wait := sync.WaitGroup{}
	for i := 0; i < threads; i++ {
		wait.Add(1)
		go func() {
			for {
				if len(ch) == 0 {
					break
				}

				if v, ok := <-ch; ok {
					// Move funds
					executeCC(client, v)
					value := queryCC(client, v)
					if value == nil {
						fmt.Println("value EMPTY")
					}
					//fmt.Println("value is ", str)

				} else {
					break
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
	t := time.Since(start)

	return fmt.Sprintf("%s%s \n%s%d \n%s%d \n%s%s", "Testing start time: ", start.String(), "Threads: ", threads, "Rounds: ", rounds, "Takes time: ", t), nil
}

func executeCC(client *channel.Client, i int) {
	response, err := client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "set", Args: getTxArgs(i)},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil || response.TxValidationCode != 0 {
		fmt.Println(i, " execute: ", string(response.Payload), response.TransactionID, response.TxValidationCode, response.ChaincodeStatus)
	}
	//fmt.Println(i, " execute: ", string(response.Payload), response.TransactionID, response.TxValidationCode, response.ChaincodeStatus)
}

func queryCC(client *channel.Client, i int) []byte {
	response, err := client.Query(channel.Request{ChaincodeID: ccID, Fcn: "get", Args: getQueryArgs(i)},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		fmt.Printf("Failed to query : %s\n", err)
		fmt.Println(i, " query: ", string(response.Payload), response.TransactionID, response.TxValidationCode, response.ChaincodeStatus)
	}
	fmt.Println(i, " query: ", string(response.Payload), response.TransactionID, response.TxValidationCode, response.ChaincodeStatus)
	return response.Payload
}

func stress_test_main() {
	configPath := "./network-config.yaml"
	//End to End testing
	setupAndRun(config.FromFile(configPath), 0, 0)
}

func StressTest(threads, rounds int) (string, error) {
	configPath := "./network-config.yaml"
	return setupAndRun(config.FromFile(configPath), threads, rounds)
}
