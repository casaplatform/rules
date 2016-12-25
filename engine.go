package rules

import (
	"github.com/casaplatform/casa"
	"github.com/casaplatform/casa/cmd/casa/environment"
	"github.com/casaplatform/mqtt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Register will add the rule to the gloval RuleList. This allows you to choose
// which rules are compiled in by using "import _ " statements
func Register(rule Rule) {
	RuleList = append(RuleList, rule)
}

// Gloval rule list
var RuleList []Rule

// Rule defines an automation rule
type Rule interface {
	Name() string
	Topics() []string
	HandleMessage(topic string, payload []byte, ch chan casa.Message) error
}

// Engine processes rules and forwards messages to and from them.
type Engine struct {
	Rules []Rule
	ch    chan casa.Message
	done  chan struct{}

	casa.Logger
	casa.MessageClient
}

// Add rule to the Engine
func (e *Engine) Register(rule Rule) {
	e.Rules = append(e.Rules, rule)
}

// Set the logger to use for output
func (e *Engine) UseLogger(logger casa.Logger) {
	e.Logger = logger
}

// Register this service with the environment global services list.
func init() {
	environment.RegisterService("rules", &Engine{})
}

// Start the rules service.
func (e *Engine) Start(config *viper.Viper) error {
	ch := make(chan casa.Message)
	done := make(chan struct{})
	e.done = done

	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				e.PublishMessage(msg)
			}
		}
	}()

	handler := func(msg *casa.Message, err error) {
		if err != nil {
			e.Log("ERROR: " + err.Error())
		}

		// Pass the message to each rule
		for i := 0; i < len(e.Rules); i++ {
			r := e.Rules[i]
			// Use a go routine so other rules aren't held up by
			// a single slow one.
			go func() {
				err = r.HandleMessage(msg.Topic, msg.Payload, ch)
				if err != nil {
					e.Log("Error executing rule '"+r.Name()+"':", err)
				}
			}()
		}
	}

	var userString string
	if config.IsSet("MQTT.User") {
		userString = config.GetString("MQTT.User") + ":" +
			config.GetString("MQTT.Pass") + "@"
	}

	client, err := mqtt.NewClient("tcp://" + userString + "127.0.0.1:1883")
	if err != nil {
		return errors.Wrap(err, "Unable to create client")
	}

	e.MessageClient = client
	e.Handle(handler)

	// Pull in the rules that were registered with init() functions
	for _, r := range RuleList {
		e.Register(r)
	}

	// Subscribe to all the topics that rules need
	for _, r := range e.Rules {
		for _, t := range r.Topics() {
			err = e.MessageClient.Subscribe(t)
			if err != nil {
				return errors.Wrap(err, "Unable to subscribe to topic")
			}
		}
	}
	return nil
}

// Stop the rules service
func (e *Engine) Stop() error {
	e.done <- struct{}{}
	return e.MessageClient.Close()
}
