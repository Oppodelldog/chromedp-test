package runner

import (
	"context"
	"fmt"
	"os"
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
}

// Suites runs the given suites.
func Suites(url string, suites TestSuites, opts Options) {
	suiteNames := getExecutionSuiteNames(suites, opts)

	for _, suiteName := range suiteNames {
		suite := suites[suiteName]

		fmt.Println()
		fmt.Println()
		fmt.Println("----------------------------------------------------")
		fmt.Printf("Test runSuite: %s\n", suiteName)
		fmt.Println("----------------------------------------------------")
		runSuite(url, suite, opts)
	}
}

func runSuite(url string, suite TestSuite, opts Options) {
	testStartTime := time.Now()
	s := 0
	f := 0

	alloCtx, cancelAllocator := getAllocator()
	defer cancelAllocator()

	ctx, cancel := chromedp.NewContext(alloCtx)
	defer cancel()

	results := make(testResults, len(suite))
	testNames := getExecutionTestNames(suite, opts)

	for _, testName := range testNames {
		testCase := suite[testName]

		fmt.Println("----------------------------------------------------")
		fmt.Printf("Case: %s\n", testName)
		fmt.Println("----------------------------------------------------")

		results.Start(testName)

		err := testCase(ctx, url)
		if err != nil {
			f++

			results.End(testName, false, err)
			takeFailureScreenshot(ctx, testName, err)
		} else {
			s++
			results.End(testName, true, nil)
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

	if len(results.GetFailed()) > 0 {
		os.Exit(1)
	}
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
