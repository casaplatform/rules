package rules

import (
	"github.com/casaplatform/casa"
	"github.com/casaplatform/casa/cmd/casa/environment"
	"github.com/casaplatform/mqtt"
	"github.com/spf13/viper"
)

func Register(rule Rule) {
	RuleList = append(RuleList, rule)
}

var RuleList []Rule

type Rule interface {
	Name() string
	Topics() []string
	HandleMessage(topic string, payload []byte) error
}

type Engine struct {
	Rules []Rule

	casa.Logger
	casa.MessageClient
}

func (e *Engine) Register(rule Rule) {
	e.Rules = append(e.Rules, rule)
}

func (e *Engine) UseLogger(logger casa.Logger) {
	e.Logger = logger
}
func init() {
	environment.RegisterService("rules", &Engine{})

}
func (e *Engine) Start(config *viper.Viper) error {
	// Pull in the rules that were registered with init() functions
	for _, r := range RuleList {
		e.Register(r)
	}

	handler := func(msg *casa.Message, err error) {
		if err != nil {
			e.Log("ERROR: " + err.Error())
		}

		// Pass the message to each rule
		for _, r := range e.Rules {
			err = r.HandleMessage(msg.Topic, msg.Payload)
			if err != nil {
				e.Log("Error executing rule '"+r.Name()+"':", err)
			}
		}
		//e.Log("Executed rule '" + r.Name() + "'")
	}

	var userString string
	if config.IsSet("MQTT.User") {
		userString = config.GetString("MQTT.User") + ":" +
			config.GetString("MQTT.Pass") + "@"
	}

	client, err := mqtt.NewClient("tcp://" + userString + "127.0.0.1:1883")
	if err != nil {
		return err
	}

	e.MessageClient = client
	e.Handle(handler)

	// Subscribe to all the topics that rules need
	for _, r := range e.Rules {
		for _, t := range r.Topics() {
			err = e.MessageClient.Subscribe(t)
			if err != nil {
				e.Log(err)
				return err
			}
			e.Log("Rules sub'd to", t)
		}
	}
	return nil
}

func (e *Engine) Stop() error {
	return e.MessageClient.Close()
}
