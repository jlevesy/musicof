package slack

import (
	"errors"
	"log"

	"github.com/nlopes/slack"
)

// Bot is a bot, it can be stopped
type Bot interface {
	Stop() error
}

type rtmBot struct {
	slack   *slack.Client
	rtm     *slack.RTM
	channel *slack.Channel

	halt chan chan error

	logger *log.Logger
}

// NewRTMBot builds an RTM bot
func NewRTMBot(token, channelID string, logger *log.Logger) (Bot, error) {
	client := slack.New(token)

	channel, err := client.GetChannelInfo(channelID)

	if err != nil {
		return nil, err
	}

	logger.Println("Starting the musicof game in :", channel.Name)

	rtm := client.NewRTM()
	go rtm.ManageConnection()

	bot := rtmBot{
		slack:   client,
		rtm:     rtm,
		channel: channel,
		halt:    make(chan chan error),
		logger:  logger,
	}

	go bot.loop()

	return &bot, nil

}

func (r *rtmBot) Stop() error {
	res := make(chan error)
	r.halt <- res
	return <-res
}

func (r *rtmBot) loop() {
	var err error

	for {
		select {
		case evt := <-r.rtm.IncomingEvents:
			err = r.handleEvent(evt)
		case res := <-r.halt:
			r.handleHalt()
			res <- err
			return
		}

	}
}

func (r *rtmBot) handleEvent(msg slack.RTMEvent) error {
	switch ev := msg.Data.(type) {
	case *slack.ConnectingEvent:
		r.logger.Println("Connecting...", ev.Attempt)
	case *slack.ConnectionErrorEvent:
		return ev
	case *slack.InvalidAuthEvent:
		return errors.New("Invalid auth received")
	case *slack.HelloEvent:
		r.logger.Println("Received hello, sending greetings !")
		r.rtm.SendMessage(r.rtm.NewOutgoingMessage("Hello, I'm musicof, let's play !", r.channel.ID))
	case *slack.ConnectedEvent:
		r.logger.Println("Connected !")
	case *slack.MessageEvent:
		if err := r.handleMessage(ev); err != nil {
			return err
		}
	}

	return nil

}

func (r *rtmBot) handleHalt() {
	r.rtm.SendMessage(r.rtm.NewOutgoingMessage("See you later !", r.channel.ID))
}

func (r *rtmBot) handleMessage(ev *slack.MessageEvent) error {
	// TODO write something interesting

	return nil
}