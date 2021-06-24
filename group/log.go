package group

import (
	"context"
	"github.com/Oppodelldog/chromedp-test"
	"github.com/chromedp/chromedp"
)

// Text writes a text to the output log.
func Text(text string) chromedp.Action {
	return logAction{
		text: text,
	}
}

type logAction struct {
	text string
}

func (e logAction) Do(_ context.Context) error {
	chromedptest.Printf("%v\n", e.text)

	return nil
}
