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
	testContext := runner.MustGetTestContext(ctx)
	testContext.GroupName = g.title
	runner.SetTestContext(ctx, testContext)

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
	screenshotOptions := testContext.ScreenshotOptions
	if screenshotOptions.BeforeGroup {
		testContext.ActionName = "before"
		runner.SetTestContext(ctx, testContext)
		err = runner.Screenshot(screenshotFilename(s.title, testContext, "1-before")).Do(newCtx)
		if err != nil {
			return err
		}
	}

	for _, action := range s.actions {
		testContext, newCtx := increaseTestContextStep(ctx)
		testContext.ActionName = fmt.Sprintf("before %T", action)
		runner.SetTestContext(ctx, testContext)

		if screenshotOptions.BeforeAction {
			err = runner.Screenshot(screenshotFilename(s.title, testContext, testContext.ActionName+"-before")).Do(newCtx)
			if err != nil {
				fmt.Println("err in screenshot: ", err.Error())
				return nil
			}
		}

		err = action.Do(newCtx)
		if err != nil {
			break
		}

		if screenshotOptions.AfterAction {
			testContext.ActionName = fmt.Sprintf("after %T", action)
			runner.SetTestContext(ctx, testContext)
			err = runner.Screenshot(screenshotFilename(s.title, testContext, testContext.ActionName+"-before")).Do(newCtx)
			if err != nil {
				fmt.Println("err in screenshot: ", err.Error())
				return nil
			}
		}
	}

	if screenshotOptions.AfterGroup {
		testContext.ActionName = "after"
		runner.SetTestContext(ctx, testContext)
		err = runner.Screenshot(screenshotFilename(s.title, testContext, "2-after")).Do(newCtx)
		if err != nil {
			fmt.Println("err in screenshot: ", err.Error())
			return nil
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
	testContext := runner.MustGetTestContext(ctx)
	testContext.TestStep++
	ctx = runner.SetTestContext(ctx, testContext)

	return testContext, ctx
}
