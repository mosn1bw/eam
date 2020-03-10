package main

import (
	"log"

	"github.com/yaiio/ea-messenger/subscription"
)

func main2() {
	srv := subscription.NewSubscriptionService("","")

	sub := subscription.NewSubscriptionData(
		"U4960c75d28849705bca861ff06c70f2f32",
		"Nguan2",
		"http://dl.profile.line-cdn.net/0m01a3e9e87251ad61583d39ddfc1dbaeb195ff0168af7",
		"subscribed",
	)

	err := srv.SubscribeMember(sub)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
