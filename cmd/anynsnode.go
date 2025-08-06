package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/anyproto/any-ns-node/account"
	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	"github.com/anyproto/any-ns-node/alchemysdk"
	"github.com/anyproto/any-ns-node/anynsaarpc"
	"github.com/anyproto/any-ns-node/anynsrpc"
	"github.com/anyproto/any-ns-node/cache"
	mongo "github.com/anyproto/any-ns-node/db"
	"github.com/anyproto/any-ns-node/nonce_manager"
	"github.com/anyproto/any-ns-node/queue"
	"github.com/getsentry/sentry-go"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/metric"
	nsclient "github.com/anyproto/any-sync/nameservice/nameserviceclient"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/util/crypto"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"

	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/nodeconfsource"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/nodeconf/nodeconfstore"

	"github.com/anyproto/any-sync/net/peerservice"
	"github.com/anyproto/any-sync/net/pool"
	"github.com/anyproto/any-sync/net/rpc/limiter"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/net/secureservice"
	"github.com/anyproto/any-sync/net/transport/quic"
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
	flagClient     = flag.Bool("cl", false, "run nsp client")
	command        = flag.String("cmd", "", "command to run: [admin-name-register, admin-name-renew, admin-fund-user, is-name-available, name-by-address, get-operation, batch-is-name-available, batch-name-by-anyid, name-by-anyid]")
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

	// init Sentry
	if conf.Sentry.Dsn != "" {
		err = sentry.Init(sentry.ClientOptions{
			Dsn:   conf.Sentry.Dsn,
			Debug: true,
			// capture 100% of messages
			TracesSampleRate: 1.0,
			Release:          app.Version(),
			Environment:      conf.Sentry.Environment,
		})
		if err != nil {
			log.Fatal("sentry.Init", zap.Error(err))
		}

		sentry.CaptureMessage("It works!")
	}

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
	var client = a.MustComponent(nsclient.CName).(nsclient.AnyNsClientService)

	// check commands
	switch *command {
	case "admin-name-register":
		adminNameRegister(ctx, a, client)
	case "admin-name-renew":
		adminNameRenew(ctx, a, client)
	case "is-name-available":
		clientIsNameAvailable(ctx, client)
	case "batch-is-name-available":
		clientBatchIsNameAvailable(ctx, client)
	// TODO:
	//case "batch-name-by-address":
	case "batch-name-by-any-id":
		clientBatchGetNameByAnyId(ctx, client)
	case "name-by-address":
		clientNameByAddress(ctx, client)
	case "name-by-anyid":
		clientNameByAnyid(ctx, client)
	// hidden command
	case "benchmark":
		clientBenchmark(ctx, client)

	// AccountAbstraction methods:
	case "get-user-account":
		clientGetUserAccount(ctx, client)
	case "admin-fund-user":
		// it will pack and sign the request
		// no need to do that manually
		adminFundUserAccount(ctx, a, client)
	case "get-operation":
		clientGetOperation(ctx, client)
	default:
		log.Fatal("unknown command", zap.String("command", *command))
	}
}

func clientIsNameAvailable(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.NameAvailableRequest{}
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

func clientBatchIsNameAvailable(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.BatchNameAvailableRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.BatchIsNameAvailable(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientBatchGetNameByAnyId(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.BatchNameByAnyIdRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.BatchGetNameByAnyId(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func adminNameRegister(ctx context.Context, a *app.App, client nsclient.AnyNsClientService) {
	var req = &nsp.NameRegisterRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	marshalled, err := req.MarshalVT()
	if err != nil {
		log.Fatal("can't marshal request", zap.Error(err))
	}

	var reqSigned = &nsp.NameRegisterRequestSigned{}
	reqSigned.Payload = marshalled

	acc := a.MustComponent("config").(accountservice.ConfigGetter).GetAccount()

	signKey, err := crypto.DecodeKeyFromString(
		acc.SigningKey,
		crypto.UnmarshalEd25519PrivateKey,
		nil)

	if err != nil {
		log.Fatal("can't read signing key", zap.Error(err))
	}

	// SignKey is used to sign the request
	sign, err := signKey.Sign(marshalled)
	if err != nil {
		log.Fatal("can't sign request", zap.Error(err))
	}
	reqSigned.Signature = sign

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.AdminRegisterName(ctx, reqSigned)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func adminNameRenew(ctx context.Context, a *app.App, client nsclient.AnyNsClientService) {
	var req = &nsp.NameRenewRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	marshalled, err := req.MarshalVT()
	if err != nil {
		log.Fatal("can't marshal request", zap.Error(err))
	}

	var reqSigned = &nsp.NameRenewRequestSigned{}
	reqSigned.Payload = marshalled

	acc := a.MustComponent("config").(accountservice.ConfigGetter).GetAccount()

	signKey, err := crypto.DecodeKeyFromString(
		acc.SigningKey,
		crypto.UnmarshalEd25519PrivateKey,
		nil)

	if err != nil {
		log.Fatal("can't read signing key", zap.Error(err))
	}

	// SignKey is used to sign the request
	sign, err := signKey.Sign(marshalled)
	if err != nil {
		log.Fatal("can't sign request", zap.Error(err))
	}
	reqSigned.Signature = sign

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.AdminRenewName(ctx, reqSigned)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

// run is-name-avail 1000 times
func clientBenchmark(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.NameAvailableRequest{}

	// calculate time difference
	start := time.Now()

	// do 1000 iterations
	for i := 0; i < 1000; i++ {
		// generate random name with .any suffix
		length := 10
		b := make([]byte, length+2)
		_, err := rand.Read(b)
		if err != nil {
			log.Fatal("can't generate random name", zap.Error(err))
		}

		req.FullName = fmt.Sprintf("%x", b)[2:length+2] + ".any"

		log.Info("sending request", zap.String("name", req.FullName))

		resp, err := client.IsNameAvailable(ctx, req)
		if err != nil {
			log.Fatal("can't get response", zap.Error(err))
		}
		log.Info("got response", zap.Bool("Available", resp.Available))
	}

	elapsed := time.Since(start)
	log.Info("Benchmark took", zap.Duration("elapsed", elapsed))
}

func clientNameByAddress(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.NameByAddressRequest{}
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

func clientNameByAnyid(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.NameByAnyIdRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	log.Info("sending request", zap.Any("request", req))

	resp, err := client.GetNameByAnyId(ctx, req)
	if err != nil {
		log.Fatal("can't get response", zap.Error(err))
	}
	log.Info("got response", zap.Any("response", resp))
}

func clientGetUserAccount(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.GetUserAccountRequest{}
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

func adminFundUserAccount(ctx context.Context, a *app.App, client nsclient.AnyNsClientService) {
	// 1 - pack request
	var req = &nsp.AdminFundUserAccountRequest{}
	err := json.Unmarshal([]byte(*params), &req)
	if err != nil {
		log.Fatal("wrong command parameters", zap.Error(err))
	}

	marshalled, err := req.MarshalVT()
	if err != nil {
		log.Fatal("can't marshal request", zap.Error(err))
	}

	var reqSigned = &nsp.AdminFundUserAccountRequestSigned{}
	reqSigned.Payload = marshalled

	acc := a.MustComponent("config").(accountservice.ConfigGetter).GetAccount()

	signKey, err := crypto.DecodeKeyFromString(
		acc.SigningKey,
		crypto.UnmarshalEd25519PrivateKey,
		nil)

	if err != nil {
		log.Fatal("can't read signing key", zap.Error(err))
	}

	// SignKey is used to sign the request
	sign, err := signKey.Sign(marshalled)
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

func clientGetOperation(ctx context.Context, client nsclient.AnyNsClientService) {
	var req = &nsp.GetOperationStatusRequest{}

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
		Register(quic.New()).
		Register(secureservice.New()).
		Register(server.New()).
		Register(nsclient.New())
}

func BootstrapServer(a *app.App) {
	a.Register(account.New()).
		Register(contracts.New()).
		Register(metric.New()).
		Register(nodeconf.New()).
		Register(nodeconfstore.New()).
		Register(nodeconfsource.New()).
		Register(coordinatorclient.New()).
		Register(alchemysdk.New()).
		Register(limiter.New()).
		Register(cache.New()).
		Register(pool.New()).
		Register(peerservice.New()).
		Register(secureservice.New()).
		Register(server.New()).
		Register(accountabstraction.New()).
		Register(anynsrpc.New()).
		Register(anynsaarpc.New()).
		Register(queue.New()).
		Register(mongo.New()).
		Register(nonce_manager.New()).
		Register(yamux.New()).
		Register(quic.New())
}
