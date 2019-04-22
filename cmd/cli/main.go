package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/matsilva/dsm-sync/pkg/invision"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	inv      *invision.Invision
	out      string
}

func validFlag(f, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s\tThis flag is cannot be blank", f)
	}
	return nil
}

func main() {
	//Parse cmd line flags
	assetURL := flag.String("asset-url", "https://cloudbees.invisionapp.com/dsm/cloudbees/cbds/applications/data-export/less", "URL where invision dsm assets can be found")
	userName := flag.String("u", "", "Invision username")
	pass := flag.String("p", "", "Invision password")
	out := flag.String("o", "./", "Folder to place downloaded asset(s)")
	flag.Parse()

	//Setup custom logs
	errorLog := log.New(os.Stderr, "ERROR\t ", log.LUTC|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stderr, "INFO\t ", log.LUTC|log.Ltime|log.Lshortfile)

	//Validate cmd line flags
	err := validFlag("-u", *userName)
	if err != nil {
		errorLog.Fatal(err)
		return
	}
	err = validFlag("-p", *pass)
	if err != nil {
		errorLog.Fatal(err)
		return
	}

	//Setup application
	inv := &invision.Invision{
		UserName: *userName,
		Pass:     *pass,
		AssetURL: *assetURL,
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		inv:      inv,
		out:      *out,
	}

	//start chrome driver

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)

	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var res string

	err = chromedp.Run(ctx,
		//Login
		chromedp.Navigate(app.inv.AssetURL),
		chromedp.WaitVisible(`#emailAddress`),
		chromedp.SendKeys(`#emailAddress`, app.inv.UserName),
		chromedp.SendKeys(`#password`, app.inv.Pass),
		chromedp.Click(`button`, chromedp.ByQuery),
		//wait for download link after redirect
		chromedp.WaitVisible(`a[download]`, chromedp.ByQuery), //currently hangs here
		chromedp.InnerHTML(`body`, &res, chromedp.ByQuery),
	)
	// TODO: display error for incorrect username/pass

	if err != nil {
		app.errorLog.Fatal(err)
	}
	infoLog.Println(res)
}
