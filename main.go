package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetClient struct {
	service       *sheets.Service
	spreadSheetId string
	sheetName     string
}

var cellsMap = map[string]string{
	"Tha":         "C",
	"khovovo":     "D",
	"TCastellani": "E",
}

var months = map[time.Month]string{
	time.January:   "Janeiro",
	time.February:  "Fevereiro",
	time.March:     "Março",
	time.April:     "Abril",
	time.May:       "Maio",
	time.June:      "Junho",
	time.July:      "Julho",
	time.August:    "Agosto",
	time.September: "Setembro",
	time.October:   "Outubro",
	time.November:  "Novembro",
	time.December:  "Dezembro",
}

// var days = map[time.Weekday]string{
// 	time.Sunday:    "Dom",
// 	time.Monday:    "Seg",
// 	time.Tuesday:   "Ter",
// 	time.Wednesday: "Qua",
// 	time.Thursday:  "Qui",
// 	time.Friday:    "Sex",
// 	time.Saturday:  "Sáb",
// }

var (
	currentSheet *SheetClient
)

func NewSheetClient(credentialsFile, spreadSheetId, sheetName string) (*SheetClient, error) {
	ctx := context.Background()

	service, err := sheets.NewService(ctx,
		option.WithCredentialsFile(credentialsFile))

	if err != nil {
		return nil, err
	}

	return &SheetClient{
		service:       service,
		spreadSheetId: spreadSheetId,
		sheetName:     sheetName,
	}, nil
}

func (c *SheetClient) AppendRow(cell string, values []interface{}) error {
	range_ := c.sheetName + "!" + cell

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.spreadSheetId,
		range_,
		valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	return err
}

func PresenceHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.GlobalName)
	log.Println(m.Content)

	cell := cellsMap[m.Author.GlobalName]

	if cell != "" {
		err := currentSheet.AppendRow(cell, []any{true})

		if err != nil {
			log.Println(fmt.Errorf("error while appending new row, %v", err))
		}
	} else {
		fmt.Println("User not mapped")
	}
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	spreadSheetId := os.Getenv("SHEET_ID")
	now := time.Now()
	sheetName := months[now.Month()]
	currentSheet, err = NewSheetClient(credentialsFile, spreadSheetId, sheetName)

	if err != nil {
		log.Fatal("Error creating the Sheet client")
	}

	bot := fmt.Sprintf("Bot %s", os.Getenv("BOT_TOKEN"))

	session, _ := discordgo.New(bot)

	session.AddHandler(PresenceHandler)

	if err := session.Open(); err != nil {
		log.Fatalf("Error opening connection, %s", err)

		return
	}

	defer session.Close()

	log.Println("connection opened")

	log.Println("Bot is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP)
	<-sc
}
