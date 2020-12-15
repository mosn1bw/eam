// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"log"
	"net/http"

	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/yaiio/ea-messenger/utils"
	"github.com/yaiio/ea-messenger/subscription"
)


func main() {
	app, err := NewEaBot(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
		os.Getenv("APP_BASE_URL"),
		os.Getenv("GOOGLE_JSON_KEY"),
		os.Getenv("SPREAD_SHEET_ID"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		})

	// serve /static/** files
	staticFileServer := http.FileServer(http.Dir("static"))
	http.HandleFunc("/static/", http.StripPrefix("/static/", staticFileServer).ServeHTTP)
	// serve /downloaded/** files
	downloadedFileServer := http.FileServer(http.Dir(app.downloadDir))
	http.HandleFunc("/downloaded/", http.StripPrefix("/downloaded/", downloadedFileServer).ServeHTTP)

	http.HandleFunc("/callback", app.Callback)
	// This is just a sample code.
	// For actually use, you must support HTTPS by using `ListenAndServeTLS`, reverse proxy or etc.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}


// EaBot app
type EaBot struct {
	bot        *linebot.Client
	appBaseURL string
	downloadDir string
	subscriptionService *subscription.Service
}

// EaBot function
func NewEaBot(channelSecret, channelToken, appBaseURL, googleJsonKeyBase64, spreadsheetId string) (*EaBot, error) {

	fmt.Println(fmt.Sprintf("channelSecret: %s", channelSecret))
	fmt.Println(fmt.Sprintf("channelToken: %s", channelToken))
	fmt.Println(fmt.Sprintf("appBaseURL: %s", appBaseURL))
	fmt.Println(fmt.Sprintf("spreadsheetId: %s", spreadsheetId))

	apiEndpointBase := os.Getenv("ENDPOINT_BASE")
	if apiEndpointBase == "" {
		apiEndpointBase = linebot.APIEndpointBase
	}

	bot, err := linebot.New(
		channelSecret,
		channelToken,
		linebot.WithEndpointBase(apiEndpointBase), // Usually you omit this.
	)
	if err != nil {
		return nil, err
	}
	downloadDir := filepath.Join(filepath.Dir(os.Args[0]), "line-bot")
	_, err = os.Stat(downloadDir)
	if err != nil {
		if err := os.Mkdir(downloadDir, 0777); err != nil {
			return nil, err
		}
	}
	subscriptionService := subscription.NewSubscriptionService(googleJsonKeyBase64, spreadsheetId)
	return &EaBot{
		bot:        bot,
		appBaseURL: appBaseURL,
		downloadDir: downloadDir,
		subscriptionService: subscriptionService,
	}, nil
}

// Callback function for http server
func (app *EaBot) Callback(w http.ResponseWriter, r *http.Request) {
	events, err := app.bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		log.Printf("Got event %v", event)
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if os.Getenv("HANDLE_TEXT") == "true" {
					if err := app.handleText(message, event.ReplyToken, event.Source); err != nil {
						log.Print(err)
					}
				}
			default:
				log.Printf("Unknown message: %v", message)
			}
		case linebot.EventTypeFollow:
			log.Printf("Got followed event: %v", event.Source.UserID)
			//if err := app.replyText(event.ReplyToken, "Got followed event"); err != nil {
			//	log.Print(err)
			//}

			profile, err := app.bot.GetProfile(event.Source.UserID).Do()
			if err != nil {
				log.Print( err.Error())
			}

			subData := subscription.NewSubscriptionData(
				profile.UserID,
				profile.DisplayName,
				profile.PictureURL,
				"subscribed",
			)

			log.Printf("Process sub data: %v", subData)

			err = app.subscriptionService.SubscribeMember(subData)
			if err != nil {
				if err2 := app.replyText(event.ReplyToken, err.Error()); err2 != nil {
					log.Print( err2.Error())
				}
			}

			log.Printf("Welcome: " + profile.DisplayName + ", UserID: " + event.Source.UserID)
			if err := app.replyText(event.ReplyToken, "Welcome: " + profile.DisplayName); err != nil {
				log.Print( err.Error())
			}

		case linebot.EventTypeUnfollow:
			log.Printf("Unfollowed this bot: %v", event)
		case linebot.EventTypeJoin:
			log.Printf("Join this bot via %v, mid: %v", event.Source.Type, event.Source.UserID)
			if err := app.replyText(event.ReplyToken, "Joined "+string(event.Source.Type)); err != nil {
				log.Print(err)
			}
		case linebot.EventTypeLeave:

			profile, err := app.bot.GetProfile(event.Source.UserID).Do()
			if err != nil {
				log.Print( err.Error())
			}
			log.Printf("%v Left: %v", profile.DisplayName, event)
			if _, err := app.bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage("Bye see you again: " + profile.DisplayName),
			).Do(); err != nil {
				log.Print( err.Error())
			}
		case linebot.EventTypePostback:
			data := event.Postback.Data
			if data == "DATE" || data == "TIME" || data == "DATETIME" {
				data += fmt.Sprintf("(%v)", *event.Postback.Params)
			}
			if err := app.replyText(event.ReplyToken, "Got postback: "+data); err != nil {
				log.Print(err)
			}
		default:
			log.Printf("Unknown event: %v", event)
		}
	}
}

func (app *EaBot) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	switch message.Text {
	case "profile":
		if source.UserID != "" {
			profile, err := app.bot.GetProfile(source.UserID).Do()
			if err != nil {
				return app.replyText(replyToken, err.Error())
			}
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage("Display name: "+profile.DisplayName),
				linebot.NewTextMessage("Status message: "+profile.StatusMessage),
				linebot.NewTextMessage("Group:"+source.GroupID),
			).Do(); err != nil {
				return err
			}
		} else {
			return app.replyText(replyToken, "Bot can't use profile API without user ID")
		}
	case "confirm":
		template := linebot.NewConfirmTemplate(
			"Do it?",
			linebot.NewMessageAction("Yes", "Yes!"),
			linebot.NewMessageAction("No", "No!"),
		)
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage("Confirm alt text", template),
		).Do(); err != nil {
			return err
		}

	case "approve":
		profile, err := app.bot.GetProfile(source.UserID).Do()
		if err != nil {
			log.Print( err.Error())
		}
		encodeUserId := utils.EncodeUserId(source.UserID)
		text := "\"" + profile.DisplayName + "\" request the subscription, Approve?"
		approvedText := "Approve subscriber \"" + profile.DisplayName + "\" (" + encodeUserId + ")"
		rejectedText := "Reject subscriber \"" + profile.DisplayName + "\" (" + encodeUserId + ")"
		template := linebot.NewConfirmTemplate(
			text,
			linebot.NewMessageAction("Approve", approvedText),
			linebot.NewMessageAction("Reject", rejectedText),
		)
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage(text, template),
		).Do(); err != nil {
			return err
		}
	case "bye":
		switch source.Type {
		case linebot.EventSourceTypeUser:
			return app.replyText(replyToken, "Bot can't leave from 1:1 chat. GroupID:" + source.GroupID)
		case linebot.EventSourceTypeGroup:
			if err := app.replyText(replyToken, "Leaving group" + source.GroupID); err != nil {
				return err
			}
			if _, err := app.bot.LeaveGroup(source.GroupID).Do(); err != nil {
				return app.replyText(replyToken, err.Error())
			}
		case linebot.EventSourceTypeRoom:
			if err := app.replyText(replyToken, "Leaving room: " + source.RoomID); err != nil {
				return err
			}
			if _, err := app.bot.LeaveRoom(source.RoomID).Do(); err != nil {
				return app.replyText(replyToken, err.Error())
			}
		}
	case "@all":
		switch source.Type {
		case linebot.EventSourceTypeUser:
			log.Printf("Only You!! " + source.UserID)

			profile, err := app.bot.GetProfile(source.UserID).Do()
			if (err != nil) {
				return err
			}
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage("Hello: "+profile.DisplayName),
			).Do(); err != nil {
				return err
			}

			return app.replyText(replyToken, "Only You!! " + source.UserID)
		case linebot.EventSourceTypeGroup:
			if err := app.replyText(replyToken, "All Members In group " + source.GroupID); err != nil {
				return err
			}
			continuationToken := ""
			if members, err := app.bot.GetGroupMemberIDs(source.GroupID, continuationToken).Do(); err != nil {
				return err
			} else {
				for _, userId := range members.MemberIDs {
					if err := app.replyText(replyToken, "userId " + userId); err != nil {
						return err
					}

					profile, err := app.bot.GetProfile(userId).Do()
					if (err != nil) {
						return err
					}
					if _, err := app.bot.ReplyMessage(
						replyToken,
						linebot.NewTextMessage("Display name: "+profile.DisplayName),
					).Do(); err != nil {
						return err
					}
				}
			}

		case linebot.EventSourceTypeRoom:
			if err := app.replyText(replyToken, "All Members In Room " + source.RoomID); err != nil {
				return err
			}

			log.Printf("GetRoomMemberIDs %s", source.RoomID)
			continuationToken := ""
			if members, err := app.bot.GetRoomMemberIDs(source.RoomID, continuationToken).Do(); err != nil {
				log.Printf("All Members In Room: %s", err.Error())
				return err
			} else {
				log.Printf("Echo message to %s: %s", replyToken, message.Text)
				for i := range members.MemberIDs {
					userId := members.MemberIDs[i]
					if err := app.replyText(replyToken, "userId " + userId); err != nil {
						return err
					}

					profile, err := app.bot.GetRoomMemberProfile(source.RoomID, userId).Do()
					if (err != nil) {
						return err
					}
					if _, err := app.bot.ReplyMessage(
						replyToken,
						linebot.NewTextMessage("userId: "+userId),
						linebot.NewTextMessage("Display name: "+profile.DisplayName),
					).Do(); err != nil {
						return err
					}
				}
			}
		}
	default:
		log.Printf("Echo message to %s: %s %s", replyToken, message.Text , source.UserID)

		if strings.HasPrefix(message.Text, "Approve subscriber") || strings.HasPrefix(message.Text, "Reject subscriber") {
			replyText := ""

			userId := utils.ExtractEncodeUserId(message.Text)
			profile, err := app.bot.GetProfile(userId).Do()
			if (err != nil) {
				return err
			}

			if strings.HasPrefix(message.Text, "Approve subscriber") {
				replyText = "Approved, " + profile.DisplayName
			} else if strings.HasPrefix(message.Text, "Reject subscriber") {
				replyText = "Rejected, " + profile.DisplayName
			}

			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage(replyText),
			).Do(); err != nil {
				return err
			}
		} else {
			if os.Getenv("ENABLED_ECHO") == "true" {
				if _, err := app.bot.ReplyMessage(
					replyToken,
					linebot.NewTextMessage(message.Text),
				).Do(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (app *EaBot) replyText(replyToken, text string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}
