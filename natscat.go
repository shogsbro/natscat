// Copyright 2017 Sigurd Hogsbro
// NATSCAT
// =======

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/nats-io/go-nats"
	"github.com/urfave/cli"
)

var (
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
		message = strings.Join(c.Args()[0:], " ")
		buffered = true
	}

	if subject == "" {
		log.Fatal("Must specify subject string")
	}

	if !listen {
		if strings.IndexAny(subject, "*>") >= 0 {
			log.Fatal("Cannot specify wildcard subject when publishing")
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
		log.Println("Connected to", nc.ConnectedUrl())
	}

	switch {
	case listen:
		// Listening for messages
		if verbose {
			log.Printf("Listening on [%s], buffered %v\n", subject, buffered)
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
			log.Printf("[%s] Wrote '%s'", subject, message)
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
				log.Printf("[%s] Wrote %d lines", subject, count)
			}
		} else {
			bytes, _ := ioutil.ReadAll(os.Stdin)
			count = len(bytes)
			nc.Publish(subject, bytes)
			if verbose {
				log.Printf("[%s] Wrote %d bytes", subject, count)
			}
		}
	}

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	//	reader := bufio.NewReader(os.Stdin)
	//	fmt.Println(reader.ReadLine())
}

func main() {
	// Log to stderr without timestamp
	log.SetFlags(0)

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "print the version",
	}
	app := cli.NewApp()
	app.Usage = "cat to/from NATS subject"
	app.UsageText = "natscats [global options] topic [message to post (buffered mode)]"
	app.Version = "0.1"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "buffered, b",
			Usage:       "Read/write messages in buffered mode, terminated by CR/LF",
			Destination: &buffered,
		},
		cli.StringFlag{
			Name:        "message, m",
			Usage:       "Write message",
			Value:       "",
			Destination: &message,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "verbose logging",
			Destination: &verbose,
		},
		cli.BoolFlag{
			Name:        "listen, l",
			Usage:       "listen for messages on NATS subject",
			Destination: &listen,
		},
		cli.StringFlag{
			Name:        "subject, s",
			Value:       "",
			Usage:       "NATS subject (* and > wildcards only valid when listening)",
			Destination: &subject,
		},
		cli.StringFlag{
			Name:        "server, S",
			Value:       nats.DefaultURL,
			Usage:       "NATS server URL(s), comma-separated",
			EnvVar:      "NATS",
			Destination: &serverURL,
		},
	}
	app.Action = cmdLine
	app.Run(os.Args)

	cat()
}
