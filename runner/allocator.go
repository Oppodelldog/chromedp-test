package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const chromeDebugURLTimeout = time.Second * 5

func getAllocator() (context.Context, context.CancelFunc) {
	var (
		allocCtx        context.Context
		cancel          context.CancelFunc
		remoteChromeURL = getDebugURL()
	)

	if remoteChromeURL != "" {
		allocCtx, cancel = chromedp.NewRemoteAllocator(context.Background(), remoteChromeURL)
	} else {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("ignore-certificate-errors", "1"),
		)
		allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}

	return allocCtx, cancel
}

func getDebugURL() string {
	v, ok := os.LookupEnv("REMOTE_CHROME_HOST")
	if !ok {
		return ""
	}

	fmt.Printf("REMOTE_CHROME_HOST is set: %v\n", v)
	parts := strings.SplitN(v, ":", 2)
	host := parts[0]
	port := parts[1]

	addr, err := net.LookupIP(host)
	if err != nil {
		log.Fatal("Unknown host")
	} else {
		fmt.Printf("looked up IP address for chrome: %v\n", addr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), chromeDebugURLTimeout)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr[0].String()+":"+port+"/json/version", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	var result map[string]interface{}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&result); err != nil {
		fmt.Println(string(bodyBytes))
		log.Fatal(err)
	}

	wsAddress := result["webSocketDebuggerUrl"].(string)
	fmt.Printf("got ws remote address: %v\n", wsAddress)

	return wsAddress
}
