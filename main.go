package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"

	"google.golang.org/protobuf/proto"
)

func GetEventHandler(client *whatsmeow.Client) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe == true {
				fmt.Println("Message from me. Will not process")
				return
			}

			fmt.Println("----------------------------NEW MESSAGE---------------------------")
			fmt.Println("Chat: ", v.Info.Chat)
			fmt.Println("Id: ", v.Info.ID)
			fmt.Println("Sender: ", v.Info.Sender)
			fmt.Println("Name: ", v.Info.PushName)
			fmt.Println("Time: ", v.Info.Timestamp)
			fmt.Println("----------------------------BEGIN---------------------------------")

			if v.Message.Conversation != nil {
				var msg = v.Message.GetConversation()
				fmt.Println("Message: ", msg)

				if msg == "ping" {
					client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{Conversation: proto.String("pong")})
				}
			}

			var extend = v.Message.GetExtendedTextMessage()
			if extend != nil {
				if extend.ContextInfo.QuotedMessage != nil {
					// fmt.Println("QuotedMessage", extend.ContextInfo.GetQuotedMessage())
					if extend.ContextInfo.QuotedMessage.Conversation != nil {
						fmt.Println("QuotedMessageConversation: ", extend.ContextInfo.GetQuotedMessage().GetConversation())
					}

					if extend.ContextInfo.QuotedMessage.ImageMessage != nil {
						fmt.Println("QuotedMessageImageMessage: WAImageMessage")
					}

					if extend.ContextInfo.QuotedMessage.StickerMessage != nil {
						fmt.Println("QuotedMessageStickerMessage: WAStickerMessage")
					}
				}

				if extend.Text != nil {
					fmt.Println("Text: ", extend.GetText())
				}

				if extend.Title != nil {
					fmt.Println("Title: ", extend.GetTitle())
				}
			}

			if v.Message.ReactionMessage != nil {
				// fmt.Println("Reaction: ", v.Message.GetReactionMessage(), "Text: ", v.Message.GetReactionMessage().GetText())
				fmt.Println("Reaction: ", v.Message.GetReactionMessage().GetText(), v.Message.GetReactionMessage().Key.GetID())
			}

			if v.Message.ImageMessage != nil {
				// fmt.Println("ImageMessage", v.Message.GetImageMessage())
				fmt.Println("ImageMessage: WAImageMessage")
			}

			if v.Message.StickerMessage != nil {
				// fmt.Println("StickerMessage", v.Message.GetStickerMessage())
				fmt.Println("StickerMessage: WAStickerMessage")
			}

			if v.Message.CommentMessage != nil {
				// fmt.Println("CommentMessage", v.Message.GetCommentMessage())
				fmt.Println("CommentMessage: WACommentMessage")
			}

			fmt.Println("----------------------------END-----------------------------------")
			fmt.Println()
		}
	}
}

func main() {
	container, err := sqlstore.New("sqlite3", "file:whatsmeow.db?_foreign_keys=on", nil)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	client := whatsmeow.NewClient(deviceStore, nil)
	client.AddEventHandler(GetEventHandler(client))

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()

		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
