package view

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type EventList struct {
	service        types.Service
	prevPageAction func()
}

func NewEventList(service types.Service) *EventList {
	return &EventList{
		service:        service,
		prevPageAction: func() {},
	}
}

func (el *EventList) SetPrevPageAction(action func()) *EventList {
	el.prevPageAction = action
	return el
}

func (el *EventList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(el.table(), 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			el.prevPageAction()
		}
		return event
	})

	return page
}

func (el *EventList) table() *tview.Table {
	header := []string{"CREATED AT", "MESSAGE"}

	rows := make([][]string, len(el.service.Events))
	for i, event := range el.service.Events {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			event.CreatedAt.Format(time.RFC3339),
			aws.ToString(event.Message),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {})
}
