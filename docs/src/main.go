//go:generate go run gen/godoc.go
//go:generate go fmt

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/maxence-charriere/go-app/v9/pkg/analytics"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/maxence-charriere/go-app/v9/pkg/cli"
	"github.com/maxence-charriere/go-app/v9/pkg/errors"
	"github.com/maxence-charriere/go-app/v9/pkg/logs"
	"github.com/maxence-charriere/go-app/v9/pkg/ui"
)

const (
	defaultTitle       = "A Go package for building Progressive Web Apps"
	defaultDescription = "A package for building progressive web apps (PWA) with the Go programming language (Golang) and WebAssembly (Wasm). It uses a declarative syntax that allows creating and dealing with HTML elements only by using Go, and without writing any HTML markup."
	backgroundColor    = "#2e343a"

	buyMeACoffeeURL     = "https://www.buymeacoffee.com/maxence"
	openCollectiveURL   = "https://opencollective.com/go-app"
	githubURL           = "https://github.com/maxence-charriere/go-app"
	githubSponsorURL    = "https://github.com/sponsors/maxence-charriere"
	twitterURL          = "https://twitter.com/jonhymaxoo"
	coinbaseBusinessURL = "https://commerce.coinbase.com/checkout/851320a4-35b5-41f1-897b-74dd5ee207ae"
)

type localOptions struct {
	Port int `cli:"p" env:"GOAPP_DOCS_PORT" help:"The port used by the server that serves the PWA."`
}

type githubOptions struct {
	Output string `cli:"o" env:"-" help:"The directory where static resources are saved."`
}

func main() {
	ui.BaseHPadding = 42
	ui.BlockPadding = 18
	analytics.Add(analytics.NewGoogleAnalytics())

	app.Route("/", newHomePage())
	app.Route("/getting-started", newGettingStartedPage())
	app.Route("/architecture", newArchitecturePage())
	app.Route("/reference", newReferencePage())

	app.Route("/components", newComponentsPage())
	app.Route("/declarative-syntax", newDeclarativeSyntaxPage())
	app.Route("/routing", newRoutingPage())
	app.Route("/static-resources", newStaticResourcePage())
	app.Route("/js", newJSPage())
	app.Route("/concurrency", newConcurrencyPage())
	app.Route("/seo", newSEOPage())
	app.Route("/lifecycle", newLifecyclePage())
	app.Route("/install", newInstallPage())
	app.Route("/testing", newTestingPage())
	app.Route("/actions", newActionPage())
	app.Route("/states", newStatesPage())

	app.Route("/migrate", newMigratePage())
	app.Route("/github-deploy", newGithubDeployPage())

	app.Route("/privacy-policy", newPrivacyPolicyPage())

	app.Handle(installApp, handleAppInstall)
	app.Handle(updateApp, handleAppUpdate)
	app.Handle(getMarkdown, handleGetMarkdown)
	app.Handle(getReference, handleGetReference)

	app.RunWhenOnBrowser()

	ctx, cancel := cli.ContextWithSignals(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()
	defer exit()

	localOpts := localOptions{Port: 7777}
	cli.Register("local").
		Help(`Launches a server that serves the documentation app in a local environment.`).
		Options(&localOpts)

	githubOpts := githubOptions{}
	cli.Register("github").
		Help(`Generates the required resources to run the documentation app on GitHub Pages.`).
		Options(&githubOpts)

	h := app.Handler{
		Name:        "Documentation for go-app",
		Title:       defaultTitle,
		Description: defaultDescription,
		Author:      "Maxence Charriere",
		Image:       "https://go-app.dev/web/images/go-app.png",
		Keywords: []string{
			"go-app",
			"go",
			"golang",
			"app",
			"pwa",
			"progressive web app",
			"webassembly",
			"web assembly",
			"webapp",
			"web",
			"gui",
			"ui",
			"user interface",
			"graphical user interface",
			"frontend",
			"opensource",
			"open source",
			"github",
		},
		BackgroundColor: backgroundColor,
		ThemeColor:      backgroundColor,
		LoadingLabel:    "go-app documentation",
		Scripts: []string{
			"/web/js/prism.js",
		},
		Styles: []string{
			"https://fonts.googleapis.com/css2?family=Montserrat:wght@400;500&display=swap",
			"/web/css/prism.css",
			"/web/css/docs.css",
		},
		RawHeaders: []string{
			analytics.GoogleAnalyticsHeader("G-SW4FQEM9VM"),
		},
		CacheableResources: []string{
			"/web/documents/what-is-go-app.md",
			"/web/documents/updates.md",
			"/web/documents/home.md",
			"/web/documents/home-next.md",
		},
	}

	switch cli.Load() {
	case "local":
		runLocal(ctx, &h, localOpts)

	case "github":
		generateGitHubPages(ctx, &h, githubOpts)
	}
}

func runLocal(ctx context.Context, h *app.Handler, opts localOptions) {
	app.Log(logs.New("starting go-app documentation service").
		Tag("port", opts.Port).
		Tag("version", h.Version),
	)

	s := http.Server{
		Addr:    fmt.Sprintf(":%v", opts.Port),
		Handler: h,
	}

	go func() {
		<-ctx.Done()
		s.Shutdown(context.Background())
	}()

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func generateGitHubPages(ctx context.Context, h *app.Handler, opts githubOptions) {
	if err := app.GenerateStaticWebsite(opts.Output, h); err != nil {
		panic(err)
	}
}

func exit() {
	err := recover()
	if err != nil {
		app.Log("command failed:", errors.Newf("%v", err))
		os.Exit(-1)
	}
}
