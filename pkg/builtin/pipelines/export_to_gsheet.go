package pipelines

import (
	"context"
	"errors"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type export2GSHEET[OUT any] struct {
	service       *sheets.Service
	sheetName     string
	spreadSheetId string
	sheetId       int64
}

func Export2GSHEET[OUT any](keyFilePath, spreadSheetId string, sheetId int64) (*export2GSHEET[OUT], error) {
	ctx := context.Background()

	service, err := sheets.NewService(ctx, option.WithCredentialsFile(keyFilePath))

	if err != nil {
		return nil, fmt.Errorf("Export2GSHEET: error creating a service using provided creds %w", err)
	}

	response, err := service.Spreadsheets.Get(spreadSheetId).Fields("sheets(properties(sheetId,title))").Do()

	if err != nil {
		return nil, fmt.Errorf("Export2GSHEET: error getting spreadsheet by id %s %w", spreadSheetId, err)
	}

	if response.HTTPStatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Export2GSHEET: %d status code received", response.HTTPStatusCode))
	}

	sheetName := ""

	for _, sheet := range response.Sheets {
		if sheet.Properties.SheetId == sheetId {
			sheetName = sheet.Properties.Title
			break
		}
	}

	if sheetName == "" {
		return nil, errors.New("Export2GSHEET: empty sheetname")
	}

	return &export2GSHEET[OUT]{
		service:       service,
		sheetName:     sheetName,
		spreadSheetId: spreadSheetId,
		sheetId:       sheetId,
	}, nil
}

func (p *export2GSHEET[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2GSHEET[OUT]) Close() {
}

func (p *export2GSHEET[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
	}

	row := &sheets.ValueRange{
		Values: original.RecordsFlat(),
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
