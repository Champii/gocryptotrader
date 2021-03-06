package events

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/champii/gocryptotrader/common"
	"github.com/champii/gocryptotrader/config"
	"github.com/champii/gocryptotrader/currency"
	"github.com/champii/gocryptotrader/currency/pair"
	"github.com/champii/gocryptotrader/exchanges/ticker"
)

const (
	ITEM_PRICE            = "PRICE"
	GREATER_THAN          = ">"
	GREATER_THAN_OR_EQUAL = ">="
	LESS_THAN             = "<"
	LESS_THAN_OR_EQUAL    = "<="
	IS_EQUAL              = "=="
	ACTION_SMS_NOTIFY     = "SMS"
	ACTION_CONSOLE_PRINT  = "CONSOLE_PRINT"
	ACTION_TEST           = "ACTION_TEST"
	CONFIG_PATH_TEST      = config.CONFIG_TEST_FILE
)

var (
	ErrInvalidItem      = errors.New("Invalid item.")
	ErrInvalidCondition = errors.New("Invalid conditional option.")
	ErrInvalidAction    = errors.New("Invalid action.")
	ErrExchangeDisabled = errors.New("Desired exchange is disabled.")
	ErrCurrencyInvalid  = errors.New("Invalid currency.")
)

type Event struct {
	ID             int
	Exchange       string
	Item           string
	Condition      string
	FirstCurrency  string
	SecondCurrency string
	Action         func(e *Event, t *ticker.TickerPrice) bool
	Executed       bool
}

var Events []*Event

func AddEvent(Exchange, Item, Condition, FirstCurrency, SecondCurrency string, Action func(e *Event, t *ticker.TickerPrice) bool) (int, error) {
	err := IsValidEvent(Exchange, Item, Condition, Action)
	if err != nil {
		return 0, err
	}

	if !IsValidCurrency(FirstCurrency, SecondCurrency) {
		return 0, ErrCurrencyInvalid
	}

	Event := &Event{}

	if len(Events) == 0 {
		Event.ID = 0
	} else {
		Event.ID = len(Events) + 1
	}

	Event.Exchange = Exchange
	Event.Item = Item
	Event.Condition = Condition
	Event.FirstCurrency = FirstCurrency
	Event.SecondCurrency = SecondCurrency
	Event.Action = Action
	Event.Executed = false
	Events = append(Events, Event)
	return Event.ID, nil
}

func RemoveEvent(EventID int) bool {
	for i, x := range Events {
		if x.ID == EventID {
			Events = append(Events[:i], Events[i+1:]...)
			return true
		}
	}
	return false
}

func GetEventCounter() (int, int) {
	total := len(Events)
	executed := 0

	for _, x := range Events {
		if x.Executed {
			executed++
		}
	}
	return total, executed
}

func (e *Event) ExecuteAction(tick ticker.TickerPrice) bool {
	return e.Action(e, &tick)
	// if common.StringContains(e.Action, ",") {
	// 	action := common.SplitStrings(e.Action,",")
	// 	if action[0] == ACTION_SMS_NOTIFY {
	// 		message := fmt.Sprintf("Event triggered: %s", e.EventToString())
	// 		if action[1] == "ALL" {
	// 			smsglobal.SMSSendToAll(message, config.Cfg)
	// 		} else {
	// 			smsglobal.SMSNotify(smsglobal.SMSGetNumberByName(action[1], config.Cfg.SMS), message, config.Cfg)
	// 		}
	// 	}
	// } else {
	// 	log.Printf("Event triggered: %s", e.EventToString())
	// }
	// return true
}

func (e *Event) EventToString() string {
	condition := common.SplitStrings(e.Condition, ",")
	return fmt.Sprintf("If the %s%s %s on %s is %s then %s.", e.FirstCurrency, e.SecondCurrency, e.Item, e.Exchange, condition[0]+" "+condition[1], e.Action)
}

func (e *Event) CheckCondition() bool { //Add error handling
	lastPrice := 0.00
	condition := common.SplitStrings(e.Condition, ",")
	targetPrice, _ := strconv.ParseFloat(condition[1], 64)

	ticker, err := ticker.GetTickerByExchange(e.Exchange)
	if err != nil {
		return false
	}

	t := ticker.Price[pair.CurrencyItem(e.FirstCurrency)][pair.CurrencyItem(e.SecondCurrency)]

	lastPrice = ticker.Price[pair.CurrencyItem(e.FirstCurrency)][pair.CurrencyItem(e.SecondCurrency)].Last

	if lastPrice == 0 {
		return false
	}

	switch condition[0] {
	case GREATER_THAN:
		{
			if lastPrice > targetPrice {
				return e.ExecuteAction(t)
			}
		}
	case GREATER_THAN_OR_EQUAL:
		{
			if lastPrice >= targetPrice {
				return e.ExecuteAction(t)
			}
		}
	case LESS_THAN:
		{
			if lastPrice < targetPrice {
				return e.ExecuteAction(t)
			}
		}
	case LESS_THAN_OR_EQUAL:
		{
			if lastPrice <= targetPrice {
				return e.ExecuteAction(t)
			}
		}
	case IS_EQUAL:
		{
			if lastPrice == targetPrice {
				return e.ExecuteAction(t)
			}
		}
	}
	return false
}

func IsValidEvent(Exchange, Item, Condition string, Action func(e *Event, t *ticker.TickerPrice) bool) error {
	Exchange = common.StringToUpper(Exchange)
	Item = common.StringToUpper(Item)

	configPath := ""

	if !IsValidExchange(Exchange, configPath) {
		return ErrExchangeDisabled
	}

	if !IsValidItem(Item) {
		return ErrInvalidItem
	}

	if !common.StringContains(Condition, ",") {
		return ErrInvalidCondition
	}

	condition := common.SplitStrings(Condition, ",")

	if !IsValidCondition(condition[0]) || len(condition[1]) == 0 {
		return ErrInvalidCondition
	}

	return nil
}

func CheckEvents() {
	for {
		total, executed := GetEventCounter()
		if total > 0 && executed != total {
			for _, event := range Events {
				if !event.Executed {
					success := event.CheckCondition()
					if success {
						log.Printf("Event %d triggered on %s successfully.\n", event.ID, event.Exchange)
						event.Executed = true
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func IsValidCurrency(currencies ...string) bool {
	for _, whatIsIt := range currencies {
		whatIsIt = common.StringToUpper(whatIsIt)
		if currency.IsDefaultCryptocurrency(whatIsIt) {
			return true
		}
		if currency.IsDefaultCurrency(whatIsIt) {
			return true
		}
	}
	return false
}

func IsValidExchange(Exchange, configPath string) bool {
	Exchange = common.StringToUpper(Exchange)

	cfg := config.GetConfig()
	if len(cfg.Exchanges) == 0 {
		cfg.LoadConfig(configPath)
	}

	for _, x := range cfg.Exchanges {
		if x.Name == Exchange && x.Enabled {
			return true
		}
	}
	return false
}

func IsValidCondition(Condition string) bool {
	switch Condition {
	case GREATER_THAN, GREATER_THAN_OR_EQUAL, LESS_THAN, LESS_THAN_OR_EQUAL, IS_EQUAL:
		return true
	}
	return false
}

func IsValidAction(Action string) bool {
	Action = common.StringToUpper(Action)
	switch Action {
	case ACTION_SMS_NOTIFY, ACTION_CONSOLE_PRINT, ACTION_TEST:
		return true
	}
	return false
}

func IsValidItem(Item string) bool {
	Item = common.StringToUpper(Item)
	switch Item {
	case ITEM_PRICE:
		return true
	}
	return false
}
