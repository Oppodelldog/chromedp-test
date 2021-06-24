package runner

import (
	"context"
	"errors"
	"github.com/Oppodelldog/chromedp-test"
	"sort"
	"time"

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
	Timeout    time.Duration
}

// ScreenshotOptions controls screenshot behavior.
type ScreenshotOptions struct {
	OutDir         string
	OnFailure      bool
	BeforeTestCase bool
	AfterTestCase  bool
	BeforeGroup    bool
	AfterGroup     bool
	BeforeAction   bool
	AfterAction    bool
	PostProcessing PostProcessingOptions
}

type PostProcessingOptions struct {
	OutDir       string
	RemoveImages bool
	CreateGIF    bool
}

// Suites runs the given suites.
func Suites(url string, suites TestSuites, opts Options) {
	suiteNames := getExecutionSuiteNames(suites, opts)

	for id, suiteName := range suiteNames {
		suite := suites[suiteName]

		chromedptest.Printf("\n")
		chromedptest.Printf("\n")
		chromedptest.Printf("----------------------------------------------------\n")
		chromedptest.Printf("Test runSuite: %s\n", suiteName)
		chromedptest.Printf("----------------------------------------------------\n")

		if !runSuite(id, url, suiteName, suite, opts) {
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
	ActionName        string
	TestStep          int
	ScreenshotOptions ScreenshotOptions
}

type testContextKey struct{}

func MustGetTestContext(ctx context.Context) TestContext {
	return ctx.Value(testContextKey{}).(TestContext)
}

func SetTestContextData(ctx context.Context, testContext TestContext) context.Context {
	return context.WithValue(ctx, testContextKey{}, testContext)
}

func runSuite(id int, url, suiteName string, suite TestSuite, opts Options) bool {
	testStartTime := time.Now()
	s := 0
	f := 0

	results := make(testResults, len(suite))
	testNames := getExecutionTestNames(suite, opts)

	for testIdx, testName := range testNames {
		testCase := suite[testName]

		chromedptest.Printf("----------------------------------------------------\n")
		chromedptest.Printf("Case: %s\n", testName)
		chromedptest.Printf("----------------------------------------------------\n")

		results.Start(testName)

		alloCtx, cancelAllocator := getAllocator()
		testCtx, dpCancel := chromedp.NewContext(alloCtx)

		testCtxData := TestContext{
			ID:                ((id + 1) * 1000) + (testIdx + 1),
			SuiteName:         suiteName,
			TestName:          testName,
			ScreenshotOptions: opts.Screenshot,
		}
		testCtx = SetTestContextData(testCtx, testCtxData)
		to := opts.Timeout
		if to == 0 {
			to = time.Minute
		}
		ctx, cancel := context.WithTimeout(testCtx, to)
		err := testCase(ctx, url)
		if err != nil {
			f++

			results.End(testName, false, err)
			if opts.Screenshot.OnFailure && err != nil && !errors.Is(err, context.DeadlineExceeded) {
				takeFailureScreenshot(ctx, opts.Screenshot.OutDir, testName, err)
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
		cancelAllocator()
		dpCancel()
		cancel()
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
