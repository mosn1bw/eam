package subscription

import (
	"github.com/yaiio/ea-messenger/utils"
	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
	"log"
	"time"
	"fmt"
	"errors"
)

type Service struct {
	srv *sheets.Service
	spreadsheetId string
}

func NewSubscriptionService(jsonKeyBase64 string, spreadsheetId string) *Service {
	ctx := context.Background()
	client := utils.NewClientFromJWT(ctx, jsonKeyBase64)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}

	return &Service{
		srv: srv,
		spreadsheetId: spreadsheetId,
	}
}


type SubscriptionData struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	JoinDate time.Time `json:"joinDate"`
	Status string `json:"status"`
}

func NewSubscriptionData(userId, displayName, pictureUrl, status string) *SubscriptionData {
	return &SubscriptionData{
		UserID: userId,
		DisplayName: displayName,
		PictureURL: pictureUrl,
		JoinDate: time.Now(),
		Status: status,
	}
}

func (s *SubscriptionData) ToArray() ([]interface{}) {
	return []interface{} {
		s.UserID,
		s.DisplayName,
		s.PictureURL,
		s.JoinDate,
		s.Status,
	}
}


func (s *Service) SubscribeMember(sub *SubscriptionData) error {
	log.Printf("Subscribe Member: %v ", sub)
	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1fMQDhxj7iu8A1XCOaHeEb8dTiTa6rrTh9CGv5ImKKWQ/edit
	//spreadsheetId := "1fMQDhxj7iu8A1XCOaHeEb8dTiTa6rrTh9CGv5ImKKWQ"
	readRange := "Subscriber!A1:E"

	resp, err := s.srv.Spreadsheets.Values.Get(s.spreadsheetId, readRange).Do()
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to retrieve data from sheet. %v", err))
	}

	// find existing user
	userIdColNo := 0
	if len(resp.Values) > 0 {
		// fmt.Println("User ID, Display Name, Picture URL, Join Date, Status:")
		for rowNo, row := range resp.Values {
			if rowNo == 0 {
			} else {
				// fmt.Printf("%s, %s, %s, %s, %s\n", row[0], row[1], row[2], row[3], row[4])
				if row[userIdColNo] == sub.UserID {
					return errors.New(fmt.Sprintf("%v already subscribed", sub.DisplayName))
				}
			}
		}
	} else {
		fmt.Print("No data found.")
	}

	valueRange := &sheets.ValueRange{
		Values:  [][]interface{}{
			sub.ToArray(),
		},
	}
	_, err = s.srv.Spreadsheets.Values.Append(s.spreadsheetId, readRange, valueRange).
		ValueInputOption("RAW").Do()
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to retrieve data from sheet. %v", err))
	}

	return nil
}