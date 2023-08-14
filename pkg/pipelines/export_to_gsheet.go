package pipelines

import (
	"context"
	"errors"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type export2GSHEET[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	service       *sheets.Service
	sheetName     string
	spreadSheetId string
	sheetId       int64
	onOpenHook    OpenHook
	onCloseHook   CloseHook
}

func Export2GSHEET[IN core.Job, OUT any](keyFilePath, spreadSheetId string, sheetId int64) (*export2GSHEET[IN, OUT, core.Output[IN, OUT]], error) {
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

	return &export2GSHEET[IN, OUT, core.Output[IN, OUT]]{
		service:       service,
		sheetName:     sheetName,
		spreadSheetId: spreadSheetId,
		sheetId:       sheetId,
	}, nil
}

func (p *export2GSHEET[IN, OUT, OR]) SetOpenHook(open OpenHook) *export2GSHEET[IN, OUT, OR] {
	p.onOpenHook = open
	return p
}

func (p *export2GSHEET[IN, OUT, OR]) SetCloseHook(close CloseHook) *export2GSHEET[IN, OUT, OR] {
	p.onCloseHook = close
	return p
}

func (p *export2GSHEET[IN, OUT, OR]) Open(ctx context.Context) error {
	if p.onOpenHook == nil {
		return nil
	}
	return p.onOpenHook(ctx)
}

func (p *export2GSHEET[IN, OUT, OR]) Close() {
	if p.onCloseHook == nil {
		return
	}
	p.onCloseHook()
}

func (p *export2GSHEET[IN, OUT, OR]) ProcessItem(input any, original OR, MetaData metadata.MetaData) (any, error) {

	if original.IsEmpty() {
		return nil, nil
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
		return nil, err
	}

	return nil, nil
}
