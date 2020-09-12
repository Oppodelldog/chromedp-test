package group

import (
	"context"

	"github.com/chromedp/chromedp"
)

//New creates a group of chromedp.Action that are grouped by a title.
func New(title string, action ...chromedp.Action) chromedp.Action {
	return groupAction{
		title:        title,
		simpleAction: simple(action...),
	}
}

type groupAction struct {
	title        string
	simpleAction simpleAction
}

func (g groupAction) Do(ctx context.Context) error {
	err := Text(g.title).Do(ctx)
	if err != nil {
		return err
	}

	return g.simpleAction.Do(ctx)
}

func simple(action ...chromedp.Action) simpleAction {
	return simpleAction{
		actions: action,
	}
}

type simpleAction struct {
	actions []chromedp.Action
}

func (s simpleAction) Do(ctx context.Context) error {
	var err error

	for _, action := range s.actions {
		err = action.Do(ctx)
		if err != nil {
			break
		}
	}

	return err
}
