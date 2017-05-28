package exchange

import (
	"log"
	"time"

	"github.com/champii/gocryptotrader/common"
	"github.com/champii/gocryptotrader/config"
	"github.com/champii/gocryptotrader/currency/pair"
	"github.com/champii/gocryptotrader/exchanges/orderbook"
	"github.com/champii/gocryptotrader/exchanges/ticker"
)

const (
	WarningBase64DecryptSecretKeyFailed = "WARNING -- Exchange %s unable to base64 decode secret key.. Disabling Authenticated API support."
	ErrExchangeNotFound                 = "Exchange not found in dataset."
)

//ExchangeAccountInfo : Generic type to hold each exchange's holdings in all enabled currencies
type ExchangeAccountInfo struct {
	ExchangeName string
	Currencies   []ExchangeAccountCurrencyInfo
}

//ExchangeAccountCurrencyInfo : Sub type to store currency name and value
type ExchangeAccountCurrencyInfo struct {
	CurrencyName string
	TotalValue   float64
	Hold         float64
}

type ExchangeBase struct {
	Name                        string
	Enabled                     bool
	Verbose                     bool
	Websocket                   bool
	RESTPollingDelay            time.Duration
	AuthenticatedAPISupport     bool
	APISecret, APIKey, ClientID string
	TakerFee, MakerFee, Fee     float64
	BaseCurrencies              []string
	AvailablePairs              []string
	EnabledPairs                []string
	WebsocketURL                string
	APIUrl                      string
}

type Trades struct {
	Type      string  `json:"type"`
	Price     float64 `json:"bid"`
	Amount    float64 `json:"amount"`
	TID       float64 `json:"tid"`
	Timestamp float64 `json:"timestamp"`
}

type TradeHistory struct {
	Pair      string  `json:"pair"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	OrderID   float64 `json:"order_id"`
	MyOrder   float64 `json:"is_your_order"`
	Timestamp float64 `json:"timestamp"`
}

type ActiveOrders struct {
	Pair             string  `json:"pair"`
	Type             string  `json:"sell"`
	Amount           float64 `json:"amount"`
	Rate             float64 `json:"rate"`
	TimestampCreated float64 `json:"time_created"`
	Status           float64 `json:"status"`
}

//IBotExchange : Enforces standard functions for all exchanges supported in gocryptotrader
type IBotExchange interface {
	Setup(exch config.ExchangeConfig)
	Start()
	SetDefaults()
	GetName() string
	IsEnabled() bool
	GetTickerPrice(currency pair.CurrencyPair) (ticker.TickerPrice, error)
	GetOrderbookEx(currency pair.CurrencyPair) (orderbook.OrderbookBase, error)
	GetEnabledCurrencies() []string
	GetExchangeAccountInfo() (ExchangeAccountInfo, error)

	Trade(string, string, float64, float64) (int64, error)
	GetTrades(string) ([]Trades, error)
	GetTradeHistory(int64, int64, int64, string, string, string, string) (map[string]TradeHistory, error)
	GetTradeHistory2(int64) (map[string]TradeHistory, error)
	GetActiveOrders(string) (map[string]ActiveOrders, error)
	CancelOrder(int64) (bool, error)
}

func (e *ExchangeBase) GetName() string {
	return e.Name
}
func (e *ExchangeBase) GetEnabledCurrencies() []string {
	return e.EnabledPairs
}
func (e *ExchangeBase) SetEnabled(enabled bool) {
	e.Enabled = enabled
}

func (e *ExchangeBase) IsEnabled() bool {
	return e.Enabled
}

func (e *ExchangeBase) SetAPIKeys(APIKey, APISecret, ClientID string, b64Decode bool) {
	e.APIKey = APIKey
	e.ClientID = ClientID

	if b64Decode {
		result, err := common.Base64Decode(APISecret)
		if err != nil {
			e.AuthenticatedAPISupport = false
			log.Printf(WarningBase64DecryptSecretKeyFailed, e.Name)
		}
		e.APISecret = string(result)
	} else {
		e.APISecret = APISecret
	}
}

func (e *ExchangeBase) UpdateAvailableCurrencies(exchangeProducts []string) error {
	exchangeProducts = common.SplitStrings(common.StringToUpper(common.JoinStrings(exchangeProducts, ",")), ",")
	diff := common.StringSliceDifference(e.AvailablePairs, exchangeProducts)
	if len(diff) > 0 {
		cfg := config.GetConfig()
		exch, err := cfg.GetExchangeConfig(e.Name)
		if err != nil {
			return err
		} else {
			log.Printf("%s Updating available pairs. Difference: %s.\n", e.Name, diff)
			exch.AvailablePairs = common.JoinStrings(exchangeProducts, ",")
			cfg.UpdateExchangeConfig(exch)
		}
	}
	return nil
}
