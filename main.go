package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cjongseok/mtproto"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	appVersion    = "0.0.1"
	deviceModel   = runtime.GOOS
	systemVersion = ""
	language      = "en-US"
)

var (
	apiID       = flag.Int("api_id", 0, "API ID from Telegram")
	apiHash     = flag.String("api_hash", "", "API Hash from Telegram")
	phoneNumber = flag.String("phone_number", "", "Your account phone number")
	serverIP    = flag.String("server_ip", "", "Telegram server IP")
	serverPort  = flag.Int("server_port", 443, "Telegram port IP")
	contactName = flag.String("contact_name", "", "Name of contact to take backup")
	credentials = "credentials.json"
)

func handleError(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		fmt.Fprintf(os.Stderr, "Error: %s:%d %v\n", fn, line, err)
		os.Exit(2)
	}
}

func getUsersFromMessages(msgs *mtproto.TypeMessagesMessages) []User {
	rawUsers := msgs.GetMessagesMessagesSlice().Users

	users := make([]User, len(rawUsers))
	for idx, rawUser := range rawUsers {
		users[idx] = User{
			Id:        rawUser.GetUser().Id,
			FirstName: rawUser.GetUser().FirstName,
			LastName:  rawUser.GetUser().LastName,
			Username:  rawUser.GetUser().Username,
			Phone:     rawUser.GetUser().Phone,
		}
	}

	return users
}

func getMessagesFromMessagesSlice(msgs *mtproto.TypeMessagesMessages) []Message {
	rawMessages := msgs.GetMessagesMessagesSlice().Messages

	messages := make([]Message, len(rawMessages))
	for idx, rawMessage := range rawMessages {
		msg := rawMessage.GetMessage()
		if msg == nil {
			continue
		}
		messages[idx] = Message{
			Id:        msg.Id,
			From:      msg.FromId,
			Timestamp: msg.Date,
			Message:   msg.Message,
		}
	}

	return messages
}

func userHasName(user *mtproto.PredUser, name string) bool {
	return strings.Contains(user.GetFirstName(), name) ||
		strings.Contains(user.GetLastName(), name) ||
		strings.Contains(user.GetUsername(), name)
}

func findContactByName(caller mtproto.RPCaller, name string) ([]*mtproto.PredUser, error) {
	resp, err := caller.ContactsGetContacts(context.Background(), &mtproto.ReqContactsGetContacts{})
	if err != nil {
		return nil, err
	}

	users := make([]*mtproto.PredUser, 0)
	for _, userMsg := range resp.GetContactsContacts().GetUsers() {
		user := userMsg.GetUser()
		if userHasName(user, name) {
			users = append(users, user)
		}
	}

	return users, nil
}

func getUserMessageHistory(caller mtproto.RPCaller, user *mtproto.PredUser, offset int32) (*mtproto.TypeMessagesMessages, error) {
	/*	if offset == 0 {
		offset = math.MaxInt32
	}*/
	msgs, err := caller.MessagesGetHistory(context.Background(), &mtproto.ReqMessagesGetHistory{
		Limit:     100,
		AddOffset: offset,
		Peer: &mtproto.TypeInputPeer{Value: &mtproto.TypeInputPeer_InputPeerUser{
			&mtproto.PredInputPeerUser{
				UserId:     user.Id,
				AccessHash: user.AccessHash,
			},
		},
		},
	})
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func main() {
	flag.Parse()

	if _, err := os.Stat(credentials); os.IsNotExist(err) {
		credentials = ""
	}

	config, err := mtproto.NewConfiguration(appVersion, deviceModel, systemVersion, language, 0, 0, credentials)
	handleError(err)

	var manager *mtproto.Manager
	var mconn *mtproto.Conn
	if config.KeyPath == "" {
		config.KeyPath = "credentials.json"

		var sentCode *mtproto.TypeAuthSentCode

		manager, err = mtproto.NewManager(config)
		handleError(err)

		mconn, sentCode, err = manager.NewAuthentication(*phoneNumber, int32(*apiID), *apiHash, *serverIP, *serverPort)
		handleError(err)

		time.Sleep(1 * time.Second)

		var authCode string
		fmt.Print("\n\n\n\nEnter Code:")
		fmt.Scanf("%s", &authCode)
		_, err := mconn.SignIn(*phoneNumber, authCode, sentCode.GetValue().PhoneCodeHash)
		handleError(err)
	} else {
		manager, err = mtproto.NewManager(config)
		handleError(err)
		mconn, err = manager.LoadAuthentication()
		handleError(err)
	}

	caller := mtproto.RPCaller{mconn}

	users, err := findContactByName(caller, *contactName)
	handleError(err)

	if len(users) == 0 {
		handleError(fmt.Errorf("No user named \"%s\" found!", *contactName))
	}

	history, err := getUserMessageHistory(caller, users[0], 0)
	handleError(err)

	dialogue := Dialogue{
		Count:    history.GetMessagesMessagesSlice().Count,
		Users:    getUsersFromMessages(history),
		Messages: getMessagesFromMessagesSlice(history),
	}

	//lastItem := dialogue.Messages[len(dialogue.Messages)-1]
	i := 1
	for int32(len(dialogue.Messages)) < dialogue.Count {
		offset := int32(100 * i)
		history, err = getUserMessageHistory(caller, users[0], offset)
		msgs := getMessagesFromMessagesSlice(history)
		//		lastItem = msgs[len(msgs)-1]
		dialogue.Messages = append(dialogue.Messages, msgs...)
		fmt.Printf("\n\n\n\n\n%s\n\n\n\n", i)
		i++
	}

	jsonMsgs, err := json.Marshal(dialogue)
	handleError(err)

	file, err := os.Create("negar.json")
	handleError(err)

	fmt.Printf("\n\n\n\n\n\n\n\n\n%s\n\n\n", jsonMsgs)
	fmt.Fprintf(file, "%s", jsonMsgs)
}
