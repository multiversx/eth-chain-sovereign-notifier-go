package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"

	"github.com/multiversx/eth-chain-sovereign-notifier-go/config"
	"github.com/multiversx/eth-chain-sovereign-notifier-go/factory"
)

var log = logger.GetOrCreate("eth-chain-sovereign-notifier")

const (
	configPath = "config/config.toml"

	logsPath       = "logs"
	logFilePrefix  = "eth-notifier"
	logLifeSpanSec = 432000 // 5 days
	logLifeSpanMb  = 1024   // 1 GB
)

func main() {
	app := cli.NewApp()
	app.Name = "Ethereum sovereign chain notifier"
	app.Usage = "The Ethereum Notifier is a Go-based application designed to bridge Ethereum and a MultiversX sovereign" +
		" chain by monitoring Ethereum blockchain events and relaying them to the sovereign chain in real-time." +
		" It subscribes to new block headers and specific smart contract events, correlating events with their respective" +
		" blocks, and sends structured notifications to the sovereign chain for further processing."
	app.Flags = []cli.Flag{
		logLevel,
		logSaveFile,
		disableAnsiColor,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startNotifier

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNotifier(ctx *cli.Context) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	err = initializeLogger(ctx)
	if err != nil {
		return err
	}

	var logFile closing.Closer
	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if withLogFile {
		logFile, err = createLogger()
		if err != nil {
			return err
		}
	}

	///

	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/753d0e97603e49b9863b0c770a26dbf3")
	if err != nil {
		log.LogIfError(err)
	}

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.LogIfError(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.LogIfError(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex()) // 0xbc10defa8dda384c96a17640d84de5578804945d347072e091b4e5f390ddea7f

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.LogIfError(err)
			}

			fmt.Println(block.Hash().Hex())        // 0xbc10defa8dda384c96a17640d84de5578804945d347072e091b4e5f390ddea7f
			fmt.Println(block.Number().Uint64())   // 3477413
			fmt.Println(block.Time())              // 1529525947
			fmt.Println(block.Nonce())             // 130524141876765836
			fmt.Println(len(block.Transactions())) // 7
		}
	}
	///

	wsClient, err := factory.CreateWSETHClientNotifier(cfg)
	if err != nil {
		return fmt.Errorf("cannot create sovereign notifier, error: %w", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting ws client...")

	<-interrupt
	log.Info("closing app at user's signal")

	wsClient.Close()

	if withLogFile {
		err = logFile.Close()
		log.LogIfError(err)
	}
	return nil
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := core.LoadTomlFile(&cfg, filepath)

	log.Info("loaded config", "path", configPath)

	return cfg, err
}

func initializeLogger(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	disableAnsi := ctx.GlobalBool(disableAnsiColor.Name)
	return removeANSIColorsForLoggerIfNeeded(disableAnsi)
}

func removeANSIColorsForLoggerIfNeeded(disableAnsi bool) error {
	if !disableAnsi {
		return nil
	}

	err := logger.RemoveLogObserver(os.Stdout)
	if err != nil {
		return err
	}

	return logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
}

func createLogger() (closing.Closer, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Error("error getting working directory when trying to create logger file", "error", err)
		workingDir = ""
	}

	argsLogger := file.ArgsFileLogging{
		WorkingDir:      workingDir,
		DefaultLogsPath: logsPath,
		LogFilePrefix:   logFilePrefix,
	}
	fileLogging, err := file.NewFileLogging(argsLogger)
	if err != nil {
		return nil, fmt.Errorf("%w creating log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(logLifeSpanSec), logLifeSpanMb)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}
