package gocryptotrader

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"fmt"

	"github.com/champii/gocryptotrader/common"
	"github.com/champii/gocryptotrader/config"
	"github.com/champii/gocryptotrader/events"
	"github.com/champii/gocryptotrader/exchanges"
	"github.com/champii/gocryptotrader/exchanges/anx"
	"github.com/champii/gocryptotrader/exchanges/bitfinex"
	"github.com/champii/gocryptotrader/exchanges/bitstamp"
	"github.com/champii/gocryptotrader/exchanges/btcc"
	"github.com/champii/gocryptotrader/exchanges/btce"
	"github.com/champii/gocryptotrader/exchanges/btcmarkets"
	"github.com/champii/gocryptotrader/exchanges/gdax"
	"github.com/champii/gocryptotrader/exchanges/gemini"
	"github.com/champii/gocryptotrader/exchanges/huobi"
	"github.com/champii/gocryptotrader/exchanges/itbit"
	"github.com/champii/gocryptotrader/exchanges/kraken"
	"github.com/champii/gocryptotrader/exchanges/lakebtc"
	"github.com/champii/gocryptotrader/exchanges/liqui"
	"github.com/champii/gocryptotrader/exchanges/localbitcoins"
	"github.com/champii/gocryptotrader/exchanges/okcoin"
	"github.com/champii/gocryptotrader/exchanges/poloniex"
	"github.com/champii/gocryptotrader/exchanges/ticker"
	"github.com/champii/gocryptotrader/portfolio"
	"github.com/champii/gocryptotrader/smsglobal"
)

type ExchangeMain struct {
	anx           anx.ANX
	btcc          btcc.BTCC
	bitstamp      bitstamp.Bitstamp
	bitfinex      bitfinex.Bitfinex
	btce          btce.BTCE
	btcmarkets    btcmarkets.BTCMarkets
	gdax          gdax.GDAX
	gemini        gemini.Gemini
	okcoinChina   okcoin.OKCoin
	okcoinIntl    okcoin.OKCoin
	itbit         itbit.ItBit
	lakebtc       lakebtc.LakeBTC
	liqui         liqui.Liqui
	localbitcoins localbitcoins.LocalBitcoins
	poloniex      poloniex.Poloniex
	huobi         huobi.HUOBI
	kraken        kraken.Kraken
}

type Bot struct {
	config    *config.Config
	portfolio *portfolio.PortfolioBase
	exchange  ExchangeMain
	Exchanges []exchange.IBotExchange
	tickers   []ticker.Ticker
	shutdown  chan bool
}

var bot Bot

func setupBotExchanges() {
	for _, exch := range bot.config.Exchanges {
		for i := 0; i < len(bot.Exchanges); i++ {
			if bot.Exchanges[i] != nil {
				if bot.Exchanges[i].GetName() == exch.Name {
					bot.Exchanges[i].Setup(exch)
					if bot.Exchanges[i].IsEnabled() {
						log.Printf("%s: Exchange support: %s (Authenticated API support: %s - Verbose mode: %s).\n", exch.Name, common.IsEnabled(exch.Enabled), common.IsEnabled(exch.AuthenticatedAPISupport), common.IsEnabled(exch.Verbose))
						bot.Exchanges[i].Start()
					} else {
						log.Printf("%s: Exchange support: %s\n", exch.Name, common.IsEnabled(exch.Enabled))
					}
				}
			}
		}
	}
}

func (b *Bot) Test() {
	fmt.Println("lol")
}

func Get() *Bot {
	return &bot
}

func (b *Bot) Wait() {
	<-b.shutdown
	Shutdown()
}

func (b *Bot) Start(c chan string) {
	log.SetOutput(ioutil.Discard)
	HandleInterrupt()
	b.config = &config.Cfg
	log.Printf("Loading config file %s..\n", config.CONFIG_FILE)

	err := b.config.LoadConfig("")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot '%s' started.\n", b.config.Name)
	AdjustGoMaxProcs()

	if b.config.SMS.Enabled {
		err = b.config.CheckSMSGlobalConfigValues()
		if err != nil {
			log.Println(err) // non fatal event
			b.config.SMS.Enabled = false
		} else {
			log.Printf("SMS support enabled. Number of SMS contacts %d.\n", smsglobal.GetEnabledSMSContacts(b.config.SMS))
		}
	} else {
		log.Println("SMS support disabled.")
	}

	log.Printf("Available Exchanges: %d. Enabled Exchanges: %d.\n", len(b.config.Exchanges), b.config.GetConfigEnabledExchanges())
	log.Println("Bot Exchange support:")

	b.Exchanges = []exchange.IBotExchange{
		// new(anx.ANX),
		// new(kraken.Kraken),
		// new(btcc.BTCC),
		// new(bitstamp.Bitstamp),
		// new(bitfinex.Bitfinex),
		new(btce.BTCE),
		// new(btcmarkets.BTCMarkets),
		// new(gdax.GDAX),
		// new(gemini.Gemini),
		// new(okcoin.OKCoin),
		// new(okcoin.OKCoin),
		// new(itbit.ItBit),
		// new(lakebtc.LakeBTC),
		// new(liqui.Liqui),
		// new(localbitcoins.LocalBitcoins),
		// new(poloniex.Poloniex),
		// new(huobi.HUOBI),
	}

	for i := 0; i < len(b.Exchanges); i++ {
		if b.Exchanges[i] != nil {
			b.Exchanges[i].SetDefaults()
			log.Printf("Exchange %s successfully set default settings.\n", b.Exchanges[i].GetName())
		}
	}

	setupBotExchanges()

	b.config.RetrieveConfigCurrencyPairs()

	// err = currency.SeedCurrencyData(currency.BaseCurrencies)
	// if err != nil {
	// 	log.Fatalf("Fatal error retrieving config currencies. Error: %s", err)
	// }

	log.Println("Successfully retrieved config currencies.")

	b.portfolio = &portfolio.Portfolio
	b.portfolio.SeedPortfolio(b.config.Portfolio)
	SeedExchangeAccountInfo(GetAllEnabledExchangeAccountInfo().Data)
	go portfolio.StartPortfolioWatcher()
	go func() { events.CheckEvents() }()

	// if b.config.Webserver.Enabled {
	// 	err := b.config.CheckWebserverConfigValues()
	// 	if err != nil {
	// 		log.Println(err) // non fatal event
	// 		//b.config.Webserver.Enabled = false
	// 	} else {
	// 		listenAddr := b.config.Webserver.ListenAddress
	// 		log.Printf("HTTP Webserver support enabled. Listen URL: http://%s:%d/\n", common.ExtractHost(listenAddr), common.ExtractPort(listenAddr))
	// 		router := NewRouter(b.exchanges)
	// 		log.Fatal(http.ListenAndServe(listenAddr, router))
	// 	}
	// }
	// if !b.config.Webserver.Enabled {
	// 	log.Println("HTTP Webserver support disabled.")
	// }

	// <-b.shutdown
	// Shutdown()
	c <- "ready"
}

func AdjustGoMaxProcs() {
	log.Println("Adjusting bot runtime performance..")
	maxProcsEnv := os.Getenv("GOMAXPROCS")
	maxProcs := runtime.NumCPU()
	log.Println("Number of CPU's detected:", maxProcs)

	if maxProcsEnv != "" {
		log.Println("GOMAXPROCS env =", maxProcsEnv)
		env, err := strconv.Atoi(maxProcsEnv)

		if err != nil {
			log.Println("Unable to convert GOMAXPROCS to int, using", maxProcs)
		} else {
			maxProcs = env
		}
	}
	log.Println("Set GOMAXPROCS to:", maxProcs)
	runtime.GOMAXPROCS(maxProcs)
}

func HandleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("Captured %v.", sig)
		Shutdown()
	}()
}

func Shutdown() {
	log.Println("Bot shutting down..")
	bot.config.Portfolio = portfolio.Portfolio
	err := bot.config.SaveConfig("")

	if err != nil {
		log.Println("Unable to save config.")
	} else {
		log.Println("Config file saved successfully.")
	}

	log.Println("Exiting.")
	os.Exit(1)
}

func SeedExchangeAccountInfo(data []exchange.ExchangeAccountInfo) {
	if len(data) == 0 {
		return
	}

	port := portfolio.GetPortfolio()

	for i := 0; i < len(data); i++ {
		exchangeName := data[i].ExchangeName
		for j := 0; j < len(data[i].Currencies); j++ {
			currencyName := data[i].Currencies[j].CurrencyName
			onHold := data[i].Currencies[j].Hold
			avail := data[i].Currencies[j].TotalValue
			total := onHold + avail

			if total <= 0 {
				continue
			}

			if !port.ExchangeAddressExists(exchangeName, currencyName) {
				port.Addresses = append(port.Addresses, portfolio.PortfolioAddress{Address: exchangeName, CoinType: currencyName, Balance: total, Decscription: portfolio.PORTFOLIO_ADDRESS_EXCHANGE})
			} else {
				port.UpdateExchangeAddressBalance(exchangeName, currencyName, total)
			}
		}
	}
}
