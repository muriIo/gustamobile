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

var cellsMap = map[string][]interface{}{
	"Tha":         []any{nil, nil, true, nil, nil},
	"khovovo":     []any{nil, nil, nil, true, nil},
	"TCastellani": []any{nil, nil, nil, nil, true},
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

func (c *SheetClient) AppendRow(values []interface{}) error {
	range_ := fmt.Sprintf("%s!A1:F1", c.sheetName)

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

func (c *SheetClient) checkSpreadsheet() error {
	resp, err := c.service.Spreadsheets.Get(c.spreadSheetId).Do()

	if err != nil {
		return fmt.Errorf("error accessing spreadsheet: %v", err)
	}

	fmt.Printf("Spreasheet found: %s\n", resp.Properties.Title)
	fmt.Printf("   Available Sheets:\n")

	for _, sheet := range resp.Sheets {
		fmt.Printf("       - %s (ID: %d)\n", sheet.Properties.Title, sheet.Properties.SheetId)
	}

	return nil
}

func PresenceHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.GlobalName)
	log.Println(m.Content)

	values := cellsMap[m.Author.GlobalName]

	if values != nil {
		err := currentSheet.AppendRow(values)

		if err != nil {
			log.Println(fmt.Errorf("error while appending new row, %v", err))
		}
	} else {
		fmt.Println("User not mapped")
	}
}

func main() {
	err := godotenv.Load("./config/.env")

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

	err = currentSheet.checkSpreadsheet()

	if err != nil {
		log.Fatalf("Error veryfing spreadsheet: %s", err)
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
