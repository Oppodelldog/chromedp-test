package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	chromedptest "github.com/Oppodelldog/chromedp-test"

	"github.com/chromedp/chromedp"
)

const chromeDebugURLTimeout = time.Second * 5

func getAllocator(ctx context.Context) (context.Context, context.CancelFunc) {
	var (
		allocCtx        context.Context
		cancel          context.CancelFunc
		remoteChromeURL = getDebugURL()
	)

	if remoteChromeURL != "" {
		allocCtx, cancel = chromedp.NewRemoteAllocator(ctx, remoteChromeURL)
	} else {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("ignore-certificate-errors", "1"),
		)
		allocCtx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	}

	return allocCtx, cancel
}

func getDebugURL() string {
	v, ok := os.LookupEnv("REMOTE_CHROME_HOST")
	if !ok {
		return ""
	}

	chromedptest.Printf("REMOTE_CHROME_HOST is set: %v\n", v)
	parts := strings.SplitN(v, ":", 2)
	host := parts[0]
	port := parts[1]

	addr, err := net.LookupIP(host)
	if err != nil {
		panic("unknown host")
	} else {
		chromedptest.Printf("looked up IP address for chrome: %v\n", addr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), chromeDebugURLTimeout)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr[0].String()+":"+port+"/json/version", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	var result map[string]interface{}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&result); err != nil {
		panic(err)
	}

	wsAddress, ok := result["webSocketDebuggerUrl"].(string)
	if ok {
		chromedptest.Printf("got ws remote address: %v\n", wsAddress)

		return wsAddress
	}

	return ""
}
