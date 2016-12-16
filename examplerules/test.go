package examplerules

import (
	"github.com/casaplatform/casa"
	"github.com/casaplatform/rules"
)

// A simple rule type that allows you to have multiple versions by swapping out
// the handler function.
type Tester struct {
	name    string
	topics  map[string][]byte
	handler func(topic string, payload []byte, ch chan casa.Message) error
}

// Call the handler function for the rule.
func (t Tester) HandleMessage(topic string, payload []byte, ch chan casa.Message) error {
	return t.handler(topic, payload, ch)
}

// Create and register the rules.
func init() {
	// This version of the rule only cares about a single topic
	single := Tester{
		name: "test/test = 1",
		topics: map[string][]byte{
			"test/test": nil,
		},
	}

	single.handler = single.singleTopic
	rules.Register(single)

	// This version of the rule wants the payload for two topics to be set
	// to something specific before it triggers.
	double := Tester{
		name: "test/test = 1 && test/test2 = 2",
		topics: map[string][]byte{
			"test/test":  nil,
			"test/test2": nil,
		},
	}
	double.handler = double.twoTopic
	rules.Register(double)
}

// This is the handler function for comparing the payloads of two different
// message topics.
func (t Tester) twoTopic(topic string, payload []byte, ch chan casa.Message) error {
	if topic != "test/test" && topic != "test/test2" {
		return nil
	}

	t.topics[topic] = payload

	if string(t.topics["test/test"]) == "1" &&
		string(t.topics["test/test2"]) == "2" {
		ch <- casa.Message{"Rules/test/double", []byte("triggered!"), false}
	}

	return nil
}

// This is the handler function for comparing the payload of a single topic.
func (t Tester) singleTopic(topic string, payload []byte, ch chan casa.Message) error {
	if topic != "test/test" {
		return nil
	}

	t.topics[topic] = payload

	if string(t.topics["test/test"]) == "1" {
		ch <- casa.Message{"Rules/test/single", []byte("triggered!"), false}
	}

	return nil
}

// Return the name of the rule, useful for debugging
func (t Tester) Name() string {
	return t.name
}

// Returns all the topics used in the rule so the rules service can subscribe
// to them
func (t Tester) Topics() []string {
	var out []string
	for k, _ := range t.topics {
		out = append(out, k)
	}
	return out
}
