package gocryptotrader

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/champii/gocryptotrader/currency/pair"
	"github.com/champii/gocryptotrader/exchanges/ticker"
	"github.com/gorilla/mux"
)

func jsonTickerResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	currency := vars["currency"]
	exchangeName := vars["exchangeName"]
	var response ticker.TickerPrice
	var err error
	for i := 0; i < len(bot.Exchanges); i++ {
		if bot.Exchanges[i] != nil {
			if bot.Exchanges[i].IsEnabled() && bot.Exchanges[i].GetName() == exchangeName {
				response, err = bot.Exchanges[i].GetTickerPrice(pair.NewCurrencyPairFromString(currency))
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)

	if err = encoder.Encode(response); err != nil {
		panic(err)
	}
}

type AllEnabledExchangeCurrencies struct {
	Data []EnabledExchangeCurrencies `json:"data"`
}

type EnabledExchangeCurrencies struct {
	ExchangeName   string               `json:"exchangeName"`
	ExchangeValues []ticker.TickerPrice `json:"exchangeValues"`
}

func getAllActiveTickersResponse(w http.ResponseWriter, r *http.Request) {
	var response AllEnabledExchangeCurrencies

	for _, individualBot := range bot.Exchanges {
		if individualBot != nil && individualBot.IsEnabled() {
			var individualExchange EnabledExchangeCurrencies
			individualExchange.ExchangeName = individualBot.GetName()
			log.Println("Getting enabled currencies for '" + individualBot.GetName() + "'")
			currencies := individualBot.GetEnabledCurrencies()
			log.Println(currencies)
			for _, currency := range currencies {
				tickerPrice, err := individualBot.GetTickerPrice(pair.NewCurrencyPairFromString(currency))
				if err != nil {
					continue
				}
				log.Println(tickerPrice)

				individualExchange.ExchangeValues = append(individualExchange.ExchangeValues, tickerPrice)
			}
			response.Data = append(response.Data, individualExchange)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

var ExchangeRoutes = Routes{
	Route{
		"AllActiveExchangesAndCurrencies",
		"GET",
		"/exchanges/enabled/latest/all",
		getAllActiveTickersResponse,
	},
	Route{
		"IndividualExchangeAndCurrency",
		"GET",
		"/exchanges/{exchangeName}/latest/{currency}",
		jsonTickerResponse,
	},
}
