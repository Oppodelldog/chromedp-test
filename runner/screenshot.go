package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"

	"github.com/chromedp/chromedp"
)

const quality = 90

const permDir = 0777
const permFile = 0600

func takeFailureScreenshot(ctx context.Context, dir, testID string, err error) {
	if err != nil {
		var fileName = path.Join(dir, "FAIL-"+testID+".png")

		errScreenShot := chromedp.Run(ctx, Screenshot(fileName))
		if errScreenShot != nil {
			fmt.Printf("COULD NOT TAKE SCREENSHOT: %v", errScreenShot)
		}
	}
}

// Screenshot captures a screenshot and saves it to the given filename.
func Screenshot(targetFile string) chromedp.Action {
	return takeScreenShotAction{
		targetFile: targetFile,
	}
}

type takeScreenShotAction struct {
	targetFile string
}

func (a takeScreenShotAction) Do(ctx context.Context) error {
	take(ctx, a.targetFile)

	testContext := MustGetTestContext(ctx)

	return addTitle(
		a.targetFile,
		a.targetFile,
		Title{
			SuiteName:  testContext.SuiteName,
			CaseName:   testContext.TestName,
			GroupName:  testContext.GroupName,
			ActionName: testContext.ActionName,
			Step:       testContext.TestStep,
		},
	)
}

func take(ctx context.Context, targetFile string) {
	var buf []byte

	if err := chromedp.Run(ctx, fullScreenshot(quality, &buf)); err != nil {
		log.Fatalf("error taking full screenshot: %v", err)
	}

	p := filepath.Dir(targetFile)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.Mkdir(p, permDir)
		if err != nil {
			log.Fatalf("error creating screenshot dir: %v", err)
		}
	}

	if err := ioutil.WriteFile(targetFile, buf, permFile); err != nil {
		log.Fatalf("error writing screenshot file: %v", err)
	}
}

func fullScreenshot(quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			var (
				height = int64(math.Ceil(contentSize.Height))
				width  = int64(math.Ceil(contentSize.Width))
			)

			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
	}
}
