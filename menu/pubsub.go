package menu

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"
)

func (r *REPL) handleSubscribeToTopic() error {
	p := promptui.Prompt{
		Label: "topic name",
	}
	topic, err := p.Run()
	if err != nil {
		return err
	}

	t, err := r.pubsub.Join(topic)
	if err != nil {
		return err
	}
	sub, err := t.Subscribe()
	if err != nil {
		return err
	}

	go func() {
		for {
			m, err := sub.Next(r.ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			r.m.Lock()
			msgs := r.messages[sub.Topic()]
			r.messages[sub.Topic()] = append(msgs, m)
			r.m.Unlock()
		}
	}()
	return nil
}

func (r *REPL) handlePublishToTopic() error {
	p := promptui.Prompt{
		Label: "topic name",
	}
	topic, err := p.Run()
	if err != nil {
		return err
	}

	p = promptui.Prompt{Label: "data"}
	data, err := p.Run()
	if err != nil {
		return err
	}

	t, err := r.pubsub.Join(topic)
	if err != nil {
		return err
	}
	return t.Publish(context.Background(), []byte(data))
}

func (r *REPL) handlePrintInboundMessages() error {
	r.m.RLock()
	topics := make([]string, 0, len(r.messages))
	for k := range r.messages {
		topics = append(topics, k)
	}
	r.m.RUnlock()

	s := promptui.Select{
		Label: "topic",
		Items: topics,
	}

	_, topic, err := s.Run()
	if err != nil {
		return err
	}

	r.m.Lock()
	defer r.m.Unlock()
	for _, m := range r.messages[topic] {
		fmt.Printf("<<< from: %s >>>: %s\n", m.GetFrom(), string(m.GetData()))
	}
	r.messages[topic] = nil
	return nil
}
