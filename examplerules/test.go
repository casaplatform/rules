package examplerules

import (
	"log"

	"github.com/casaplatform/rules"
)

type Tester struct {
	name    string
	topics  map[string][]byte
	handler func(topic string, payload []byte) error
}

func (t Tester) HandleMessage(topic string, payload []byte) error {
	return t.handler(topic, payload)
}

func init() {
	single := Tester{
		name: "test/test = 1",
		topics: map[string][]byte{
			"test/test": nil,
		},
	}
	single.handler = single.singleTopic

	rules.Register(single)

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
func (t Tester) twoTopic(topic string, payload []byte) error {
	if topic != "test/test" && topic != "test/test2" {
		return nil
	}

	t.topics[topic] = payload

	if string(t.topics["test/test"]) == "1" &&
		string(t.topics["test/test2"]) == "2" {
		log.Println("test/test is 1 and test/test2 is 2, executing rule!")
	}

	return nil
}

func (t Tester) singleTopic(topic string, payload []byte) error {
	if topic != "test/test" {
		return nil
	}

	t.topics[topic] = payload

	if string(t.topics["test/test"]) == "1" {
		log.Println("test/test is 1, executing rule!")
	}

	return nil
}

func (t Tester) Name() string {
	return t.name
}

func (t Tester) Topics() []string {
	var out []string
	for k, _ := range t.topics {
		out = append(out, k)
	}
	return out
}
