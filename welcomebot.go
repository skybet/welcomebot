package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

type slackResponse struct {
	Channel  string `json:"channel"`
	Raw      bool   `json:"raw_response"`
	Response string `json:"response"`
}

type Config struct {
	PublicResponses []slackResponse `json:"responses"`
	DmResponses     []slackResponse `json:"dmresponses"`
	EphResponses    []slackResponse `json:"ephresponses"`
}

var (
	botId string
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	config := loadConfig("config.json")
	api := slack.New(
		token,
		slack.OptionDebug(true),
	)

	rtm := api.NewRTM()

	go rtm.ManageConnection()

	// Return a slice of all channels from config.json
	allChans := getChannelList(config.PublicResponses, config.DmResponses, config.EphResponses)
	// Because duplicates are possible, make a new slice without duplicates
	cleanSlice := removeDuplicates(allChans)

	// Range over the cleaned up slice and join channels
	for _, v := range cleanSlice {
		_, err := api.JoinChannel(v)
		if err != nil {
			log.Errorf("Error joining public channel %s: %s", v, err)
		}
	}

Loop:
	for {

		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				botId = ev.Info.User.ID
				log.Infof("Connection counter:", ev.ConnectionCount)
			case *slack.MessageEvent:
				//only interested in public channels
				cInfo, err := api.GetChannelInfo(ev.Channel)
				if err == nil {
					if ev.SubType == "channel_join" {
						log.Infof("channel_join seen on channel: %s", ev.Msg.Channel)
						respondToJoin(rtm, ev, cInfo.Name, config)
					}
					if ev.User != botId && strings.HasPrefix(ev.Text, "<@"+botId+">") {
						log.Infof("message seen on public channel: %s", ev.Msg.Channel)
						respondToMessage(rtm, ev, cInfo.Name, config)
					}
				}
			case *slack.RTMError:
				log.Errorf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid credentials")
				break Loop

			default:
				//Take no action
			}
		}
	}
}

func sendMessage(rtm *slack.RTM, channel string, message string, raw bool) (string, string, error) {
	return rtm.PostMessage(
		channel,
		slack.MsgOptionText(message, !raw),
		slack.MsgOptionAsUser(true),
	)
}

// From https://github.com/nlopes/slack/issues/191#issuecomment-355394946
func postEphemeral(rtm *slack.RTM, channel, user, text string, raw bool) (string, error) {
	return rtm.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(text, !raw),
		slack.MsgOptionAsUser(true),
	)
}

func respondToMessage(rtm *slack.RTM, ev *slack.MessageEvent, name string, config Config) {

	acceptedGreetings := map[string]bool{
		"help": true,
	}

	text := ev.Msg.Text
	prefix := fmt.Sprintf("<@%s> ", botId)
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	if acceptedGreetings[text] {
		for _, publicResponse := range config.PublicResponses {
			if publicResponse.Channel == name {
				publicMsg := fmt.Sprintf("*Public response for this channel*:\n\n%s", publicResponse.Response)
				_, _, err := sendMessage(rtm, ev.Msg.Channel, publicMsg, publicResponse.Raw)
				if err != nil {
					log.Warnf("Error sending message: %s", err)
				}
			}
		}

		for _, dmResponse := range config.DmResponses {
			if dmResponse.Channel == name {
				dmMsg := fmt.Sprintf("*DM response for this channel*:\n\n%s", dmResponse.Response)
				_, _, err := sendMessage(rtm, ev.Msg.Channel, dmMsg, dmResponse.Raw)
				if err != nil {
					log.Warnf("Error sending message: %s", err)
				}
			}
		}

		for _, ephResponse := range config.EphResponses {
			if ephResponse.Channel == name {
				ephMsg := fmt.Sprintf("*Ephemeral response for this channel*:\n\n%s", ephResponse.Response)
				_, _, err := sendMessage(rtm, ev.Msg.Channel, ephMsg, ephResponse.Raw)
				if err != nil {
					log.Warnf("Error sending message: %s", err)
				}
			}
		}
	}
}

func respondToJoin(rtm *slack.RTM, ev *slack.MessageEvent, name string, config Config) {

	for _, publicResponse := range config.PublicResponses {
		if publicResponse.Channel == name {
			log.Infof("Sending public reply to channel %s", name)
			_, _, err := sendMessage(rtm, ev.Msg.Channel, publicResponse.Response, publicResponse.Raw)
			if err != nil {
				log.Warnf("Error sending message: %s", err)
			}
		}
	}

	for _, dmResponse := range config.DmResponses {
		if dmResponse.Channel == name {
			sta, stb, channel, err := rtm.OpenIMChannel(ev.User)
			if err != nil && sta && stb {
				log.Warnf("Failed to open IM channel to user: %s", err)
			}
			log.Infof("Sending DM to user %s", ev.User)
			_, _, err = sendMessage(rtm, channel, dmResponse.Response, dmResponse.Raw)
			if err != nil {
				log.Warnf("Error sending message: %s", err)
			}
		}
	}

	for _, ephResponse := range config.EphResponses {
		if ephResponse.Channel == name {
			log.Infof("Sending ephemeral reply to %s in channel %s", ev.User, name)
			_, err := postEphemeral(rtm, ev.Msg.Channel, ev.User, ephResponse.Response, ephResponse.Raw)
			if err != nil {
				log.Warnf("Error sending message: %s", err)
			}
		}
	}

}

func loadConfig(file string) Config {

	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Fatalf("Error opening config file %s", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func getChannelList(publicSlice, dmSlice, ephSlice []slackResponse) []string {
	var newSlice []string
	for _, v := range publicSlice {
		newSlice = append(newSlice, v.Channel)
	}
	for _, v := range dmSlice {
		newSlice = append(newSlice, v.Channel)
	}
	for _, v := range ephSlice {
		newSlice = append(newSlice, v.Channel)
	}
	return newSlice
}

func removeDuplicates(channelSlice []string) []string {
	k := make(map[string]bool)
	slice := []string{}
	for _, entry := range channelSlice {
		if _, value := k[entry]; !value {
			k[entry] = true
			slice = append(slice, entry)
		}
	}
	return slice
}
