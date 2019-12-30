// Copyright 2017 Sigurd Hogsbro
// NATSCAT
// =======

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/urfave/cli"
)

var (
	appName   = "natscat"
	listen    bool
	verbose   bool
	buffered  bool
	serverURL string
	subject   string
	message   string
)

func cmdLine(c *cli.Context) error {
	verbose = c.Bool("verbose")
	listen = c.Bool("listen")
	buffered = c.Bool("buffered")

	if !listen && c.NArg() > 0 {
		message = strings.Join(c.Args().Slice(), " ")
		buffered = true
	}

	if subject == "" {
		cli.ShowCommandHelp(c, "")
		log.Fatalf("%s: Must specify subject string\n", appName)
	}

	if !listen {
		if strings.IndexAny(subject, "*>") >= 0 {
			log.Fatalf("%s: Cannot specify wildcard subject when publishing\n", appName)
		}
	}

	return nil
}

func cat() {
	nc, err := nats.Connect(serverURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	if verbose {
		log.Printf("%s: Connected to %s\n", appName, nc.ConnectedUrl())
	}

	switch {
	case listen:
		// Listening for messages
		if verbose {
			log.Printf("%s: Listening on [%s], buffered %v\n", appName, subject, buffered)
		}
		nc.Subscribe(subject, func(m *nats.Msg) {
			if verbose {
				log.Println(m.Subject, string(m.Data))
				log.Printf("[%s] %s\n", m.Subject, string(m.Data))
			}
			if buffered {
				// print the message followed by CR/LF
				fmt.Println(string(m.Data))
			} else {
				// Write the binary message body to stdout
				buf := bytes.NewBuffer(m.Data)
				buf.WriteTo(os.Stdout)
			}
		})
		select {}

	case message != "":
		// Publish specified message
		nc.Publish(subject, []byte(message))
		if verbose {
			log.Printf("%s: [%s] Wrote '%s'\n", appName, subject, message)
		}

	case message == "":
		// Publish message(s) from stdin
		count := 0
		if buffered {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				nc.Publish(subject, []byte(line))
				count++
			}
			if verbose {
				log.Printf("%s: [%s] Wrote %d lines\n", appName, subject, count)
			}
		} else {
			bytes, _ := ioutil.ReadAll(os.Stdin)
			count = len(bytes)
			nc.Publish(subject, bytes)
			if verbose {
				log.Printf("%s: [%s] Wrote %d bytes\n", appName, subject, count)
			}
		}
	}

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Log to stderr without timestamp
	log.SetFlags(0)

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version, V",
		Usage: "print the version",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s: version %s\n", c.App.Name, c.App.Version)
		os.Exit(1)
	}
	hp := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		hp(w, templ, data)
		os.Exit(1)
	}

	app := &cli.App{
		Name:      appName,
		Usage:     "cat to/from NATS subject",
		UsageText: "natscats [global options] topic [message to post]",
		Compiled:  time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Sigurd Høgsbro",
				Email: "shogsbro@gmail.com",
			},
		},
		Version: "0.2",
		Action:  cmdLine,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "buffered, b",
				Usage:       "read/write messages in buffered mode, terminated by CR/LF",
				Destination: &buffered,
			},
			&cli.StringFlag{
				Name:        "message, m",
				Usage:       "message to publish",
				Value:       "",
				Destination: &message,
			},
			&cli.BoolFlag{
				Name:        "verbose, v",
				Usage:       "verbose logging",
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "listen, l",
				Usage:       "listen for messages",
				Destination: &listen,
			},
			&cli.StringFlag{
				Name:        "subject, s",
				Value:       "",
				Usage:       "[Required] NATS subject ('*' and '>' wildcards only valid when listening)",
				Destination: &subject,
			},
			&cli.StringFlag{
				Name:        "server, S",
				Value:       nats.DefaultURL,
				Usage:       "NATS server URL(s), comma-separated",
				EnvVars:     []string{"NATS"},
				Destination: &serverURL,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	cat()
}
