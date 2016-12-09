package nsq

import (
	"errors"
	"time"
	. "wbproject/jkglmsgpush/src/enviroment"

	"github.com/bitly/go-nsq"
)

type Handle struct {
	msgchan chan *nsq.Message
	stop    bool
}

func (h *Handle) HandleMsg(m *nsq.Message) error {
	if !h.stop {
		h.msgchan <- m
	}
	return nil
}

func (h *Handle) Process(i int) {

	h.stop = false
	for {
		select {
		case m := <-h.msgchan:

			var err error
			if i == 0 {
				err = DecodeMT(string(m.Body))
				if err != nil {
					Logger.Critical(err)
				}
			} else if i == 1 {
				err = DecodeMS(string(m.Body))
				if err != nil {
					Logger.Critical(err)
				}
			}

		case <-time.After(time.Hour):
			if h.stop {
				close(h.msgchan)
				return
			}
		}
	}
}

func (h *Handle) Stop() {
	h.stop = true
}

func NewConsummer(topic, channel string) (*nsq.Consumer, error) {

	var consumer *nsq.Consumer
	config := nsq.NewConfig()
	//心跳间隔时间 3s
	config.HeartbeatInterval = 3 * time.Second
	//3分钟去发现一次，发现topic为指定的nsqd
	config.LookupdPollInterval = 3 * time.Minute

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func ConsumerRun(consumer *nsq.Consumer, topic, address string) error {

	if consumer == nil {
		return errors.New("consumer尚未初始化 ")
	}
	if topic == "user_recipe_download" {

		h := new(Handle)
		consumer.AddHandler(nsq.HandlerFunc(h.HandleMsg))
		h.msgchan = make(chan *nsq.Message, 1024)
		err := consumer.ConnectToNSQLookupd(address)
		if err != nil {
			return err
		}
		h.Process(0)
	}

	if topic == "base_data_upload" {

		h := new(Handle)
		consumer.AddHandler(nsq.HandlerFunc(h.HandleMsg))
		h.msgchan = make(chan *nsq.Message, 1024)
		err := consumer.ConnectToNSQLookupd(address)
		if err != nil {
			return err
		}
		h.Process(1)
	}

	return nil
}
