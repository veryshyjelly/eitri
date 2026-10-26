package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/log/v2"
	"github.com/alecthomas/kong"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/server"
	"github.com/veryshyjelly/eitri/templates"
	"github.com/veryshyjelly/eitri/tui"
	"github.com/veryshyjelly/eitri/watcher"
)

// CLI represents the command-line interface arguments and flags for the application.
type CLI struct {
	Config   string `help:"Path to the configuration file." default:"./eitri.yml" short:"c"`
	Root     string `help:"Project root directory." default:"." short:"C" type:"existingdir"`
	Vertical bool   `help:"Open in vertical mode." short:"l"`
	Version  bool   `help:"Show version information." short:"v"`
	Debug    bool   `help:"Show debug information." short:"d"`
}

func main() {
	var cli CLI
	kong.Parse(&cli, kong.Name("eitri"), kong.Description("Competitive programming companion."))
	if err := run(cli); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cli CLI) error {
	if cli.Version {
		fmt.Println("eitri version 0.1.0")
		return nil
	}
	logPath := filepath.Join(os.TempDir(), "eitri.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	if cli.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	cfg, err := config.ReadConfig(cli.Config)
	if os.IsNotExist(err) {
		var language string
		err := huh.NewSelect[string]().Title("Pick a starter language configuration.").Options(
			huh.NewOption("C++", "cpp"), huh.NewOption("Rust", "rust"), huh.NewOption("Python", "python"), huh.NewOption("OCaml", "ocaml"),
		).Value(&language).Run()
		if err != nil {
			return fmt.Errorf("select language: %w", err)
		}
		data, err := templates.FS.ReadFile(language + ".yml")
		if err != nil {
			return err
		}
		f, err := os.OpenFile(cli.Config, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("create config: %w", err)
		}
		_, writeErr := f.Write(data)
		if err := errors.Join(writeErr, f.Close()); err != nil {
			return err
		}
		fmt.Printf("Created %s. Edit it for your workspace, then run eitri again.\n", cli.Config)
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	root, err := filepath.Abs(cli.Root)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg, root, cli.Config); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:6174")
	if err != nil {
		return fmt.Errorf("listen for Competitive Companion: %w", err)
	}
	defer listener.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	watch := watcher.New()
	submissions := &server.Submissions{}
	orientation := tui.Horizontal
	if cli.Vertical {
		orientation = tui.Vertical
	}
	model := tui.NewModel(ctx, root, cfg, orientation, watch, submissions)
	defer model.Close()
	p := tea.NewProgram(tui.NewSplash(model), tea.WithContext(ctx))
	intake := make(chan core.Info, 16)
	srv := server.New(listener.Addr().String(), func(requestCtx context.Context, info core.Info) error {
		deliveryCtx, stop := context.WithTimeout(requestCtx, 5*time.Second)
		defer stop()
		select {
		case intake <- info:
			return nil
		case <-deliveryCtx.Done():
			return deliveryCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}, submissions)
	var wg sync.WaitGroup
	failures := make(chan error, 1)
	wg.Add(3)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case info := <-intake:
				p.Send(info)
			}
		}
	}()
	go func() {
		defer wg.Done()
		if err := watch.Run(ctx, root, func(msg any) { p.Send(msg) }); err != nil {
			p.Send(watcher.Error{Err: err})
		}
	}()
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failures <- err
			cancel()
		}
	}()
	defer func() {
		cancel()
		shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("HTTP shutdown", "err", err)
			_ = srv.Close()
		}
		wg.Wait()
	}()
	_, err = p.Run()
	select {
	case serverErr := <-failures:
		return fmt.Errorf("HTTP server: %w", serverErr)
	default:
	}
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("terminal UI: %w", err)
	}
	return nil
}
