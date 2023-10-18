package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/anyproto/any-ns-node/account"
	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	"github.com/anyproto/any-ns-node/alchemysdk"
	"github.com/anyproto/any-ns-node/anynsaarpc"
	"github.com/anyproto/any-ns-node/anynsrpc"
	"github.com/anyproto/any-ns-node/client"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-ns-node/nonce_manager"
	as "github.com/anyproto/any-ns-node/pb/anyns_api"
	"github.com/anyproto/any-ns-node/queue"
	commonaccount "github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/util/crypto"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"

	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/nodeconfsource"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/nodeconf/nodeconfstore"

	"github.com/anyproto/any-sync/net/peerservice"
	"github.com/anyproto/any-sync/net/pool"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/net/secureservice"
	"github.com/anyproto/any-sync/net/transport/yamux"
	"go.uber.org/zap"

	// import this to keep govvv in go.mod on mod tidy
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ahmetb/govvv/integration-test/app-different-package/mypkg"
)

var log = logger.NewNamed("main")

var (
	flagConfigFile = flag.String("c", "etc/nsnode-config.yml", "path to config file")
	flagVersion    = flag.Bool("v", false, "show version and exit")
	flagHelp       = flag.Bool("h", false, "show help and exit")
	flagClient     = flag.Bool("cl", false, "run as client")
	command        = flag.String("cmd", "", "command to run: [name-register, name-renew, is-name-available, name-by-address, get-operation]")
	params         = flag.String("params", "", "command params in json format")
)

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(app.AppName)
		fmt.Println(app.Version())
		fmt.Println(app.VersionDescription())
		return
	}
	if *flagHelp {
		flag.PrintDefaults()
		return
	}

	if debug, ok := os.LookupEnv("ANYPROF"); ok && debug != "" {
		go func() {
			http.ListenAndServe(debug, nil)
		}()
	}

	// create app
	ctx := context.Background()
	a := new(app.App)

	// open config file
	conf, err := config.NewFromFile(*flagConfigFile)
	if err != nil {
		log.Fatal("can't open config file", zap.Error(err))
	}
	conf.Log.ApplyGlobal()

	// bootstrap components
	a.Register(conf)

	if *flagClient {
		runAsClient(a, ctx)
		return
	}

	BootstrapServer(a)

	// start app
	if err := a.Start(ctx); err != nil {
		log.Fatal("can't start app", zap.Error(err))
	}
	log.Info("app started", zap.String("version", a.Version()))

	// wait exit signal
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-exit
	log.Info("received exit signal, stop app...", zap.String("signal", fmt.Sprint(sig)))

	// close app
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if err := a.Close(ctx); err != nil {
		log.Fatal("close error", zap.Error(err))
	} else {
		log.Info("goodbye!")
	}
	time.Sleep(time.Second / 3)
}

func runAsClient(a *app.App, ctx context.Context) {
	log.Info("running a client...")
	BootstrapClient(a)

	// start app
	if err := a.Start(ctx); err != nil {
		log.Fatal("can't start app", zap.Error(err))
	}
	log.Info("app started", zap.String("version", a.Version()))

	// get a "client" service instance
	var client = a.MustComponent(client.CName).(client.AnyNsClientService)

	// check commands
	switch *command {
	case "name-register":
		clientNameRegister(client, ctx)
	case "is-name-available":
		clientIsNameAvailable(client, ctx)
	case "name-renew":
		clientNameRenew(client, ctx)
	case "name-by-address":
		clientNameByAddress(client, ctx)

	// AccountAbstraction methods:
	case "get-user-account":
		clientGetUserAccount(client, ctx)
	case "admin-fund-user":
		// it will pack and sign the request
		// no need to do that manually
		adminFundUserAccount(a, client, ctx)
	case "get-operation":
		clientGetOperation(client, ctx)
	default:
		log.Fatal("unknown command", zap.String("command", *command))
	}
}

func clientIsNameAvailable(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.NameAvailableRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.IsNameAvailable(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientNameRegister(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.NameRegisterRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.NameRegister(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientNameRenew(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.NameRenewRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.NameRenew(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientNameByAddress(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.NameByAddressRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.GetNameByAddress(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientGetUserAccount(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.GetUserAccountRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.GetUserAccount(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func adminFundUserAccount(a *app.App, client client.AnyNsClientService, ctx context.Context) {
	// 1 - pack request
	var req = &as.AdminFundUserAccountRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	marshalled, err := req.Marshal()
	if err != nil {
		log.Fatal("can't marshal request", zap.Error(err))
	}

	var reqSigned = &as.AdminFundUserAccountRequestSigned{}
	reqSigned.Payload = marshalled

	acc := a.MustComponent("config").(commonaccount.ConfigGetter).GetAccount()

	signingKey, err := crypto.DecodeKeyFromString(
		acc.PeerKey,
		crypto.UnmarshalEd25519PrivateKey,
		nil)

	if err != nil {
		log.Fatal("can't read signing key", zap.Error(err))
	}

	sign, err := signingKey.Sign(marshalled)
	if err != nil {
		log.Fatal("can't sign request", zap.Error(err))
	}
	reqSigned.Signature = sign

	// 3 - call
	resp, err := client.AdminFundUserAccount(ctx, reqSigned)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientGetOperation(client client.AnyNsClientService, ctx context.Context) {
	var req = &as.GetOperationStatusRequest{}

	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.GetOperation(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func BootstrapClient(a *app.App) {
	a.Register(account.New()).
		Register(nodeconf.New()).
		Register(nodeconfstore.New()).
		Register(nodeconfsource.New()).
		Register(coordinatorclient.New()).
		Register(pool.New()).
		Register(peerservice.New()).
		Register(yamux.New()).
		Register(secureservice.New()).
		Register(server.New()).
		Register(client.New())
}

func BootstrapServer(a *app.App) {
	a.Register(account.New()).
		Register(contracts.New()).
		Register(nonce_manager.New()).
		Register(nodeconf.New()).
		Register(nodeconfstore.New()).
		Register(nodeconfsource.New()).
		Register(coordinatorclient.New()).
		Register(queue.New()).
		Register(alchemysdk.New()).
		Register(pool.New()).
		Register(peerservice.New()).
		Register(yamux.New()).
		Register(secureservice.New()).
		Register(server.New()).
		Register(accountabstraction.New()).
		Register(anynsrpc.New()).
		Register(anynsaarpc.New())
}
