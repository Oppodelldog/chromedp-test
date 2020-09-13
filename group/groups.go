package group

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Oppodelldog/chromedp-test/runner"
	"github.com/chromedp/chromedp"
)

// New creates a group of chromedp.Action that are grouped by a title.
func New(title string, action ...chromedp.Action) chromedp.Action {
	return groupAction{
		title:        title,
		simpleAction: screenshot(title, action...),
	}
}

type groupAction struct {
	title        string
	simpleAction chromedp.Action
}

func (g groupAction) Do(ctx context.Context) error {
	err := Text(g.title).Do(ctx)
	if err != nil {
		return err
	}

	return g.simpleAction.Do(ctx)
}

func screenshot(title string, action ...chromedp.Action) screenshotAction {
	return screenshotAction{
		title:   title,
		actions: action,
	}
}

type screenshotAction struct {
	title   string
	actions []chromedp.Action
}

func (s screenshotAction) Do(ctx context.Context) error {
	var err error

	testContext, newCtx := increaseTestContextStep(ctx)
	if testContext.ScreenshotOptions.BeforeGroup {
		err = runner.Screenshot(screenshotFilename(s.title, testContext, "1-before")).Do(newCtx)
		if err != nil {
			return err
		}
	}

	for _, action := range s.actions {
		err = action.Do(newCtx)
		if err != nil {
			break
		}
	}

	if testContext.ScreenshotOptions.AfterGroup {
		err = runner.Screenshot(screenshotFilename(s.title, testContext, "2-after")).Do(newCtx)
		if err != nil {
			return err
		}
	}

	return err
}

func screenshotFilename(title string, testContext runner.TestContext, postfix string) string {
	return path.Join(
		testContext.ScreenshotOptions.OutDir,
		normalizeFilename(fmt.Sprintf("%v__%s_%s_%v-%s-%s.png",
			testContext.ID,
			testContext.SuiteName,
			testContext.TestName,
			testContext.TestStep,
			title,
			postfix),
		),
	)
}

func normalizeFilename(x string) string {
	var y strings.Builder

	for _, r := range x {
		if func(r rune) bool {
			switch {
			case r == '.' || r == '-' || r == '_':
				return true
			case '0' <= r && r <= '9':
				return true
			case 'a' <= r && r <= 'z':
				return true
			case 'A' <= r && r <= 'Z':
				return true
			}

			return false
		}(r) {
			y.WriteRune(r)
		}
	}

	return y.String()
}

func increaseTestContextStep(ctx context.Context) (runner.TestContext, context.Context) {
	testContext := runner.GetTestContext(ctx)
	testContext.TestStep++
	ctx = runner.SetTestContext(ctx, testContext)

	return testContext, ctx
}
