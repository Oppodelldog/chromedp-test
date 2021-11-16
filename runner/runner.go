package runner

import (
	"context"
	"sort"
	"time"

	"github.com/chromedp/cdproto/runtime"

	chromedptest "github.com/Oppodelldog/chromedp-test"
	"github.com/chromedp/chromedp"
)

// TestSuites is a dictionary of TestSuite where the key is the  test runSuite name.
type TestSuites map[string]TestSuite

// TestSuite is a dictionary of TestCase where the key is the test case name.
type TestSuite map[string]TestCase

// TestCase defines a function that runs a test.
type TestCase func(ctx context.Context, url string) error

// Options enabled to configure the run.
type Options struct {
	SortSuites bool
	SortTests  bool
	Screenshot ScreenshotOptions
	LogError   bool
	LogWarning bool
	LogInfo    bool
	LogConsole bool
}

// ScreenshotOptions controls screenshot behavior.
type ScreenshotOptions struct {
	OutDir         string
	OnFailure      bool
	PostProcessing PostProcessingOptions
}

type PostProcessingOptions struct {
	OutDir       string
	RemoveImages bool
	CreateGIF    bool
}

// Suites runs the given suites.
func Suites(ctx context.Context, url string, suites TestSuites, opts Options) {
	suiteNames := getExecutionSuiteNames(suites, opts)

	for id, suiteName := range suiteNames {
		suite := suites[suiteName]

		chromedptest.Printf("\n")
		chromedptest.Printf("\n")
		chromedptest.Printf("----------------------------------------------------\n")
		chromedptest.Printf("Test runSuite: %s\n", suiteName)
		chromedptest.Printf("----------------------------------------------------\n")

		if !runSuite(ctx, id, url, suiteName, suite, opts) {
			chromedptest.Printf("suite failed, aborting\n")

			break
		}
	}
}

type TestContext struct {
	ID                int
	SuiteName         string
	TestName          string
	GroupName         string
	ScreenshotOptions ScreenshotOptions
	Error             string
}

type testContextKey struct{}

func MustGetTestContext(ctx context.Context) TestContext {
	return ctx.Value(testContextKey{}).(TestContext)
}

func SetTestContextData(ctx context.Context, testContext TestContext) context.Context {
	return context.WithValue(ctx, testContextKey{}, testContext)
}

func runSuite(ctx context.Context, id int, url, suiteName string, suite TestSuite, opts Options) bool {
	testStartTime := time.Now()
	s := 0
	f := 0

	results := make(testResults, len(suite))
	testNames := getExecutionTestNames(suite, opts)

	alloCtx, cancelAllocator := getAllocator(ctx)
	defer cancelAllocator()

	var contextOpts []chromedp.ContextOption

	if opts.LogError {
		contextOpts = append(
			contextOpts,
			chromedp.WithDebugf(func(s string, i ...interface{}) { chromedptest.Printf(s+"\n", i...) }),
		)
	}

	if opts.LogWarning {
		contextOpts = append(
			contextOpts,
			chromedp.WithLogf(func(s string, i ...interface{}) { chromedptest.Printf(s+"\n", i...) }),
		)
	}

	if opts.LogInfo {
		contextOpts = append(
			contextOpts,
			chromedp.WithErrorf(func(s string, i ...interface{}) { chromedptest.Printf(s+"\n", i...) }),
		)
	}

	testCtx, dpCancel := chromedp.NewContext(alloCtx, contextOpts...)
	defer dpCancel()

	if opts.LogConsole {
		chromedp.ListenTarget(testCtx, func(ev interface{}) {
			if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				chromedptest.Printf("console.%s call: ", ev.Type)
				for _, arg := range ev.Args {
					chromedptest.Printf("%s - %s | ", arg.Type, arg.Value)
				}
				chromedptest.Printf("\n")
			}
		})
	}

	for testIdx, testName := range testNames {
		testCase := suite[testName]

		chromedptest.Printf("----------------------------------------------------\n")
		chromedptest.Printf("Case: %s\n", testName)
		chromedptest.Printf("----------------------------------------------------\n")

		results.Start(testName)

		testCtxData := TestContext{
			ID:                ((id + 1) * 1000) + (testIdx + 1),
			SuiteName:         suiteName,
			TestName:          testName,
			ScreenshotOptions: opts.Screenshot,
		}
		testCtx = SetTestContextData(testCtx, testCtxData)

		err := testCase(testCtx, url)
		if err != nil {
			f++

			results.End(testName, false, err)

			if opts.Screenshot.OnFailure && err != nil {
				testCtxData.Error = err.Error()
				testCtx = SetTestContextData(testCtx, testCtxData)
				takeFailureScreenshot(testCtx, opts.Screenshot.OutDir, testName, err)
			}
		} else {
			s++
			results.End(testName, true, nil)

			if opts.Screenshot.PostProcessing.CreateGIF {
				chromedptest.Printf("Creating gif\n")
				createGIF(testCtxData.ID, testCtxData.SuiteName, testCtxData.TestName, opts.Screenshot)
			}
		}

		chromedptest.Printf("----------------------------------------------------\n")

		if !results.GetSuccess(testName) {
			chromedptest.Printf("Error   : %v\n", err)
		}

		chromedptest.Printf("Success : %v\n", results.GetSuccess(testName))
		chromedptest.Printf("Duration: %v\n", results.GetDuration(testName))
		chromedptest.Printf("----------------------------------------------------\n")
	}

	chromedptest.Printf("----------------------------------------------------\n")
	chromedptest.Printf("Duration: %v\n", time.Since(testStartTime))
	chromedptest.Printf("SUCCESS : %v\n", s)
	chromedptest.Printf("FAIL    : %v\n", f)

	results.GetFailed().PrintErrors()

	return len(results.GetFailed()) == 0
}

func getExecutionSuiteNames(suites TestSuites, opts Options) []string {
	var suiteNames = make([]string, 0, len(suites))
	for k := range suites {
		suiteNames = append(suiteNames, k)
	}

	if opts.SortSuites {
		sort.Strings(suiteNames)
	}

	return suiteNames
}

func getExecutionTestNames(suite TestSuite, opts Options) []string {
	var testNames = make([]string, 0, len(suite))
	for k := range suite {
		testNames = append(testNames, k)
	}

	if opts.SortTests {
		sort.Strings(testNames)
	}

	return testNames
}
