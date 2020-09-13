package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
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
}

// ScreenshotOptions controls screenshot behavior.
type ScreenshotOptions struct {
	OutDir         string
	OnFailure      bool
	BeforeTestCase bool
	AfterTestCase  bool
	BeforeGroup    bool
	AfterGroup     bool
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
	wg := &sync.WaitGroup{}

	for id, suiteName := range suiteNames {
		suite := suites[suiteName]

		fmt.Println()
		fmt.Println()
		fmt.Println("----------------------------------------------------")
		fmt.Printf("Test runSuite: %s\n", suiteName)
		fmt.Println("----------------------------------------------------")

		if !runSuite(id, url, suiteName, suite, opts, wg) {
			fmt.Println("suite failed, aborting")

			break
		}
	}

	fmt.Println("waiting for goroutines to finish")
	wg.Wait()
	fmt.Println("goroutines finished")
}

type TestContext struct {
	ID                int
	SuiteName         string
	TestName          string
	TestStep          int
	ScreenshotOptions ScreenshotOptions
}

type testContextKey struct{}

func GetTestContext(ctx context.Context) TestContext {
	return ctx.Value(testContextKey{}).(TestContext)
}

func SetTestContext(ctx context.Context, testContext TestContext) context.Context {
	return context.WithValue(ctx, testContextKey{}, testContext)
}

func runSuite(id int, url, suiteName string, suite TestSuite, opts Options, wg *sync.WaitGroup) bool {
	testStartTime := time.Now()
	s := 0
	f := 0

	alloCtx, cancelAllocator := getAllocator()
	defer cancelAllocator()

	ctx, cancel := chromedp.NewContext(alloCtx)
	defer cancel()

	results := make(testResults, len(suite))
	testNames := getExecutionTestNames(suite, opts)

	for testIdx, testName := range testNames {
		testCase := suite[testName]

		fmt.Println("----------------------------------------------------")
		fmt.Printf("Case: %s\n", testName)
		fmt.Println("----------------------------------------------------")

		results.Start(testName)

		testCtx := TestContext{
			ID:                ((id + 1) * 1000) + (testIdx + 1),
			SuiteName:         suiteName,
			TestName:          testName,
			ScreenshotOptions: opts.Screenshot,
		}
		ctx = SetTestContext(ctx, testCtx)

		err := testCase(ctx, url)
		if err != nil {
			f++

			results.End(testName, false, err)

			if opts.Screenshot.OnFailure {
				takeFailureScreenshot(ctx, opts.Screenshot.OutDir, testName, err)
			}
		} else {
			s++
			results.End(testName, true, nil)

			if opts.Screenshot.PostProcessing.CreateGIF {
				fmt.Println("Creating gif")
				wg.Add(1)
				go func() {
					createGIF(testCtx.ID, testCtx.SuiteName, testCtx.TestName, opts.Screenshot)
					wg.Done()
				}()
			}
		}

		fmt.Println("----------------------------------------------------")

		if !results.GetSuccess(testName) {
			fmt.Printf("Error   : %v\n", err)
		}

		fmt.Printf("Success : %v\n", results.GetSuccess(testName))
		fmt.Printf("Duration: %v\n", results.GetDuration(testName))
		fmt.Println("----------------------------------------------------")
	}

	fmt.Println("----------------------------------------------------")
	fmt.Printf("Duration: %v\n", time.Since(testStartTime))
	fmt.Printf("SUCCESS : %v\n", s)
	fmt.Printf("FAIL    : %v\n", f)

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
