# chromedp-test
a small testrunner lib for chromedp

Organize ChromeDP action as Functional Web-TestCases in TestSuites and run them.

## Sample
### Organize and run Tests in TestSuites
```go
func RunTests(url string) {
	runner.Suites(url,
		runner.TestSuites{
			"1 Login Test": runner.TestSuite{
				"01-Login":        Case01Login,
			},
		},
		runner.Options{
			SortSuites: true,
			SortTests:  true,
		},
	)
}
```

### Implement a test
```go
func Case01Login(ctx context.Context, url string) error {
	return Run(ctx,
		group.New("preparations",
            NavigateToWebsite(url),
			group.New("regular login",
				Login(),
				WaitVisible(idEntryList, ByTestId),
			),
			group.New("logout to get to back logout page",
				Logout(),
			),
		),
		group.New("login from logout page",
			WaitVisible(idActionLogin, ByTestId),
			Click(idActionLogin, ByTestId),
			Login(),
		),
		group.New("expect to be logged in",
			WaitVisible(idEntryList, ByTestId),
			Logout(),
		),
	)
}
```

### Scope
This library provides runner to organize and run the tests, grouping for better logging.  
Writing ChromeDp Actions like *NavigateToWebsite* or *Login* is of course up to you.
