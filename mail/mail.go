package mail

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"get.cutie.cafe/isabelle/config"
	"get.cutie.cafe/isabelle/crash"
	"get.cutie.cafe/isabelle/discord"
	"github.com/DusanKasan/parsemail"
	"github.com/bytbox/go-pop3"
	"github.com/carlescere/scheduler"
)

// Schedule email tasks to run.
func Schedule() {
	scheduler.Every(config.Config.Email.Every).Seconds().Run(Check)
}

// Check checks the connected email address for emails with crash reports
func Check() {
	conn, err := pop3.DialTLS(config.Config.Email.Host)
	if err != nil {
		fmt.Printf("Could not dial %s: %v\n", config.Config.Email.Host, err)
		return
	}

	defer conn.Quit()

	if err = conn.Auth(config.Config.Email.User, config.Config.Email.Pass); err != nil {
		fmt.Printf("Could not auth %s: %v\n", config.Config.Email.User, err)
		return
	}

	messages, _, err := conn.ListAll()
	if err != nil {
		fmt.Printf("Could not retrieve messages: %v\n", err)
		return
	}

	for _, msgid := range messages {
		blob, err := conn.Retr(msgid)
		if err != nil {
			fmt.Printf("Could not retrieve message %d: %v\n", msgid, err)
			continue
		}

		mail, err := parsemail.Parse(strings.NewReader(blob))
		if err != nil {
			fmt.Printf("Could not parse mail %d: %v\n", msgid, err)
			continue
		}

		fmt.Printf("Received mail with subject \"%s\" from <%s> (%d, %d attachments)\n", mail.Subject, mail.From[0].Address, msgid, len(mail.Attachments))

		if len(mail.Attachments) > 0 {
			zipBytes, err := ioutil.ReadAll(mail.Attachments[0].Data)
			if err != nil {
				fmt.Printf("Could not parse attachment in mail %d: %v\n", msgid, err)
				continue
			}

			zipByteReader := bytes.NewReader(zipBytes)

			zipReader, err := zip.NewReader(zipByteReader, int64(zipByteReader.Len()))
			if err != nil {
				fmt.Printf("Could not parse ZIP in mail %d: %v\n", msgid, err)
				continue
			}

			exception, specs, err := crash.Parse(zipReader)
			if err != nil {
				fmt.Printf("Could not parse crash in mail %d\n", msgid)
				continue
			}

			discord.PostMessage(zipByteReader, mail.TextBody, "`"+mail.From[0].Address+"`", exception, specs)

			if !config.Config.Email.Keep {
				if err := conn.Dele(msgid); err != nil {
					fmt.Printf("Could not delete mail %d\n", msgid)
					continue
				}
			}
		}
	}
}
