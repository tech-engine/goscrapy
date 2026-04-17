package gsheet

import (
	"context"
	"log"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type export2GSheet[OUT any] struct {
	service       *sheets.Service
	sheetName     string
	spreadSheetId string
	sheetId       int64
}

func New[OUT any](keyFilePath, spreadSheetId string, sheetId int64) *export2GSheet[OUT] {
	ctx := context.Background()

	service, err := sheets.NewService(ctx, option.WithCredentialsFile(keyFilePath))

	if err != nil {
		log.Printf("Export2GSHEET: error creating a service using provided creds %s", err)
		return nil
	}

	response, err := service.Spreadsheets.Get(spreadSheetId).Fields("sheets(properties(sheetId,title))").Do()

	if err != nil {
		log.Printf("Export2GSHEET: error getting spreadsheet by id %s %s", spreadSheetId, err)
		return nil
	}

	if response.HTTPStatusCode != 200 {
		log.Printf("Export2GSHEET: %d status code received", response.HTTPStatusCode)
		return nil
	}

	sheetName := ""

	for _, sheet := range response.Sheets {
		if sheet.Properties.SheetId == sheetId {
			sheetName = sheet.Properties.Title
			break
		}
	}

	if sheetName == "" {
		log.Printf("Export2GSHEET: %d status code received", response.HTTPStatusCode)
		return nil
	}

	return &export2GSheet[OUT]{
		service:       service,
		sheetName:     sheetName,
		spreadSheetId: spreadSheetId,
		sheetId:       sheetId,
	}
}

func (p *export2GSheet[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2GSheet[OUT]) Close() {
}

func (p *export2GSheet[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	records := original.RecordFlat()
	row := &sheets.ValueRange{
		Values: [][]any{records},
	}

	response, err := p.service.Spreadsheets.Values.Append(p.spreadSheetId, p.sheetName, row).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(context.Background()).
		Do()

	if err != nil || response.HTTPStatusCode != 200 {
		return err
	}

	return nil
}
