package group

import (
	"context"
	chromedptest "github.com/Oppodelldog/chromedp-test"
	"time"

	"github.com/chromedp/chromedp"
)

type listAction struct {
	title   string
	timeout time.Duration
	actions []chromedp.Action
}

// New creates a group of chromedp.Action that are grouped by actions title.
func New(title string, action ...chromedp.Action) chromedp.Action {
	return newActions(0, title, action)
}

// WithTimeout creates a group of chromedp.Action where each action has the given timeout to succeed.
func WithTimeout(timeout time.Duration, title string, action ...chromedp.Action) chromedp.Action {
	return newActions(timeout, title, action)
}

func newActions(timeout time.Duration, title string, action []chromedp.Action) chromedp.Action {
	return listAction{
		title:   title,
		timeout: timeout,
		actions: action,
	}
}

func (a listAction) Do(ctx context.Context) error {
	var err error

	chromedptest.Printf("%s\n", a.title)

	for i := range a.actions {
		if a.timeout > 0 {
			err = doWithTimeout(ctx, a.timeout, a.actions[i])
		} else {
			err = a.actions[i].Do(ctx)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func doWithTimeout(ctx context.Context, timeout time.Duration, a chromedp.Action) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return a.Do(ctxTimeout)
}
