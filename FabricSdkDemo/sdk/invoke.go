package sdk

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	channelID      = "ambrchannel"
	orgName        = "org1"
	orgAdmin       = "Admin"
	ordererOrgName = "orderer.ambr.com"
	ccID           = "mycc"
)

type SdkKnife struct {
	FabSdk          *fabsdk.FabricSDK
	FabClient       *channel.Client
	ChannelProvider context.ChannelProvider
	LedgerProvider  *ledger.Client
}

func NewSdkKnife(configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) (*SdkKnife, error) {
	sdkknife := &SdkKnife{}

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		return nil, err
	}
	sdkknife.FabSdk = sdk
	// ************ setup complete ************** //
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	//clientChannelContext := sdk.NewClient(fabsdk.WithUser(orgAdmin)).Channel(channelID)
	sdkknife.ChannelProvider = clientChannelContext
	ledgerclient, ex := ledger.New(clientChannelContext)
	if ex != nil {
		return nil, ex
	}
	sdkknife.LedgerProvider = ledgerclient

	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err2 := channel.New(clientChannelContext)

	if err2 != nil {
		fmt.Printf("Failed to create new channel client: %s\n", err)
		sdk.Close()
		return nil, err2
	}
	sdkknife.FabClient = client
	return sdkknife, nil
}

func (p *SdkKnife) Dispose() {
	if p == nil {
		return
	}
	if p.FabSdk != nil {
		p.FabSdk.Close()
	}
}

func (p *SdkKnife) Invoke(ccid string, fcn string, args []string) (*channel.Response, error) {
	request := channel.Request{
		ChaincodeID: ccID,
		Fcn:         fcn,
		Args:        strings2bytes(args),
	}
	response, err := p.FabClient.Execute(request,
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		fmt.Printf("Failed to invoke : %s\n", err)
		return nil, err
	}

	return &response, nil
}

func strings2bytes(args []string) [][]byte {
	if args == nil {
		return nil
	}
	bytes := [][]byte{}
	for _, s := range args {
		bytes = append(bytes, []byte(s))
	}

	return bytes
}

func (p *SdkKnife) Query(ccid string, fcn string, args []string) (*channel.Response, error) {
	request := channel.Request{
		ChaincodeID: ccID,
		Fcn:         fcn,
		Args:        strings2bytes(args),
	}
	response, err := p.FabClient.Query(request,
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		fmt.Printf("Failed to query : %s\n", err)
		return nil, err
	}

	return &response, nil
}
