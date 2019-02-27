package main

import (
	"FabricSdkDemo/handler"
	"FabricSdkDemo/sdk"
	"fmt"
	"net/http"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)

type FabHttpHandler struct {
	Mux             *http.ServeMux
	SdkKnife        *sdk.SdkKnife
	ChannelProvider context.ChannelProvider
}

func NewFabHttpHandler() (*FabHttpHandler, error) {
	handler := &FabHttpHandler{
		Mux: http.NewServeMux(),
	}
	e := handler.init()
	if e != nil {
		return nil, e
	}

	return handler, nil
}

func (p *FabHttpHandler) Dispose() {
	if p.SdkKnife != nil {
		p.SdkKnife.Dispose()
	}
}

func (p *FabHttpHandler) InitSdk() error {
	configPath := "./network-config.yaml"
	//End to End testing
	sdk, err := sdk.NewSdkKnife(config.FromFile(configPath))
	if err != nil {
		return err
	}
	p.SdkKnife = sdk
	return nil
}

func (p *FabHttpHandler) init() error {
	e := p.InitSdk()
	if e != nil {
		return e
	}

	fs := http.FileServer(http.Dir("static"))
	p.Mux.Handle("/static/", http.StripPrefix("/static/", fs))

	ctx := &handler.HandlerContext{
		SdkKnife: p.SdkKnife,
	}

	//dynamic page
	p.Mux.HandleFunc("/about", ctx.AboutHandler)
	p.Mux.HandleFunc("/query", ctx.QueryHandler)
	p.Mux.HandleFunc("/invoke", ctx.InvokeHandler)
	p.Mux.HandleFunc("/doinvoke", ctx.DoInvokeHandler)
	p.Mux.HandleFunc("/doquery", ctx.DoQueryHandler)
	p.Mux.HandleFunc("/dotester", ctx.DoTesterHandler)
	p.Mux.HandleFunc("/register", ctx.RegisterHandler)
	p.Mux.HandleFunc("/transaction", ctx.TransactionHandler)
	p.Mux.HandleFunc("/querytransaction", ctx.TransactionHandler)

	p.Mux.HandleFunc("/channels", ctx.ChannelsHandler)
	p.Mux.HandleFunc("/tester", ctx.TesterHandler)
	p.Mux.HandleFunc("/", ctx.IndexHandler)
	return nil
}

func (p *FabHttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.Mux.ServeHTTP(w, req)
}

func Run() {
	handler, e := NewFabHttpHandler()
	if e != nil {
		fmt.Println(e)
		return
	}

	defer handler.Dispose()

	http.ListenAndServe(":8011", handler)
}

func main() {
	Run()
}
