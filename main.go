package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DusanKasan/parsemail"
	"github.com/bwmarrin/discordgo"
	"github.com/bytbox/go-pop3"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

var (
	config  configuration
	discord *discordgo.Session
	gh      *github.Client
	msgs    map[string]string
	ghctx   context.Context
)

type configuration struct {
	Discord  configurationDiscord `json:"discord"`
	Email    configurationEmail   `json:"email"`
	GitHub   configurationGitHub  `json:"github"`
	Messages _messages            `json:"messages"`
	Files    files                `json:"files"`
}

type configurationGitHub struct {
	PAT    string   `json:"pat"`
	Repo   string   `json:"repo"`
	Labels []string `json:"labels"`
	React  string   `json:"react"`
}

type configurationEmail struct {
	Host  string `json:"host"`
	User  string `json:"user"`
	Pass  string `json:"pass"`
	Every int    `json:"every"`
}

type configurationDiscord struct {
	Token  string `json:"token"`
	Output string `json:"output"`
}

type serializableException struct {
	Message    string
	Source     string
	StackTrace string
	TargetSite string
	Type       string
}

type _messages struct {
	InvalidFile string
	BadZIP      string
	BadDownload string
	Help        string
	Success     string
}

type files struct {
	File1 string `json:"file1"`
	File2 string `json:"file2"`
}

func main() {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	if err = json.Unmarshal(bytes, &config); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	discord, err = discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		fmt.Printf("Error connecting to Discord: %v\n", err)
		return
	}

	discord.AddHandler(handleDM)
	discord.AddHandler(handleReact)

	if err = discord.Open(); err != nil {
		fmt.Printf("Error connecting to Discord: %v\n", err)
		return
	}

	fmt.Printf("Connected to Discord as %s\n", discord.State.User.String())

	ghctx = context.Background()

	gh = github.NewClient(oauth2.NewClient(ghctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GitHub.PAT},
	)))

	msgs = make(map[string]string)

	killEmail := make(chan bool)
	go emailRoutine(killEmail)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Received a close signal, shutting down...")
	close(killEmail)
	discord.Close()
}

func parseCrash(zipReader *zip.Reader) (string, string, error) {
	var exception *serializableException
	var specs string

	excFmt := ""

	for _, file := range zipReader.File {
		if file.Name == config.Files.File1 || file.Name == config.Files.File2 {
			reader, err := file.Open()
			if err != nil {
				fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
				return "", "", err
			}
			defer reader.Close()

			fileBytes, err := ioutil.ReadAll(reader)
			if err != nil {
				fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
				return "", "", err
			}

			if file.Name == config.Files.File1 {
				if strings.HasSuffix(config.Files.File1, ".xml") {
					if err = xml.Unmarshal(fileBytes, &exception); err != nil {
						fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
						return "", "", err
					}
				} else {
					excFmt = string(fileBytes)
				}
			} else if file.Name == config.Files.File2 {
				specs = string(fileBytes)
			}
		}
	}

	if exception != nil {
		excFmt = exception.Type + ": " + exception.Message + "\n" + exception.StackTrace
	}

	if len(excFmt) > 1500 {
		excFmt = excFmt[0:1500]

		// make traces slightly prettier
		st := strings.Split(excFmt, "\n")
		excFmt = strings.Join(st[0:len(st)-1], "\n")
	}

	return excFmt, specs, nil
}

func reply(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	s.ChannelMessageSend(m.ChannelID, msg)
}

func handleDM(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		fmt.Printf("Unexpected error fetching channel %s: %v\n", m.ChannelID, err)
		return
	}

	if channel.Type != discordgo.ChannelTypeDM {
		return
	}

	if len(m.Attachments) == 0 || len(m.Attachments) > 1 {
		reply(s, m, fmt.Sprintf(config.Messages.Help, m.Author.Mention()))
		return
	}

	if !strings.HasSuffix(m.Attachments[0].Filename, ".zip") {
		reply(s, m, config.Messages.InvalidFile)
		return
	}

	fmt.Printf("DM received from %s#%s (%s) containing %s\n", m.Author.Username, m.Author.Discriminator, m.Author.ID, m.Attachments[0].URL)

	resp, err := http.Get(m.Attachments[0].URL)
	if err != nil {
		reply(s, m, config.Messages.BadDownload)
		fmt.Printf("Could not obtain uploaded file %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	defer resp.Body.Close()

	// zips cannot be streamed
	zipBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reply(s, m, config.Messages.BadDownload)
		fmt.Printf("Could not read uploaded file stream from %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	byteReader := bytes.NewReader(zipBytes)

	zipReader, err := zip.NewReader(byteReader, int64(byteReader.Len()))
	if err != nil {
		reply(s, m, config.Messages.BadZIP)
		fmt.Printf("Could not obtain valid zip from %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	exception, specs, err := parseCrash(zipReader)
	if err != nil {
		reply(s, m, config.Messages.BadZIP)
		return
	}

	if len(exception) == 0 || len(specs) == 0 {
		reply(s, m, config.Messages.BadZIP)
		fmt.Printf("%s was missing some files\n", m.Attachments[0].URL)
		return
	}

	postMessage(byteReader, m.Content, m.Author.Mention(), exception, specs)

	reply(s, m, config.Messages.Success)
}

func handleReact(s *discordgo.Session, react *discordgo.MessageReactionAdd) {
	if react.Emoji.Name != config.GitHub.React {
		return
	}

	msg, ok := msgs[react.MessageID]
	if !ok {
		return
	}

	issueTitle := "Crash report: " + strings.Split(msg, "\n")[2]
	issueBody := "Issue reported via Isabelle: https://discord.com/channels/" + react.MessageReaction.GuildID + "/" + config.Discord.Output + "/" + react.MessageID + "/\n\n" + strings.Join(strings.Split(msg, "\n")[1:], "\n")

	issue := &github.IssueRequest{
		Title:  &issueTitle,
		Body:   &issueBody,
		Labels: &config.GitHub.Labels}

	iss, _, err := gh.Issues.Create(ghctx, strings.Split(config.GitHub.Repo, "/")[0], strings.Split(config.GitHub.Repo, "/")[1], issue)
	if err != nil {
		s.ChannelMessageSend(react.ChannelID, fmt.Sprintf("Could not open issue: `%v`", err))
		return
	}

	_, err = s.ChannelMessageSend(react.ChannelID, fmt.Sprintf("<@"+react.UserID+">: Opened issue <%s> for `%s`", *iss.URL, issueTitle))
	if err != nil {
		fmt.Printf("Could not send message: %v\n", err)
		return
	}
}

func postMessage(byteReader *bytes.Reader, userComment string, userFrom string, exception string, specs string) (err error) {
	if len(userComment) > 0 {
		if len(userComment) > 140 {
			userComment = userComment[0:140]
		}
		userComment = "```\n" + userComment + "\n```"
	}

	if len(specs) > 300 {
		specs = specs[0:300]
	}

	message := &discordgo.MessageSend{}
	file := &discordgo.File{}

	file.ContentType = "application/zip"
	file.Name = "Report.zip"
	file.Reader = byteReader

	message.Content = fmt.Sprintf("**New bug report from %s**\n```\n%s\n```\n```\n%s\n```\n%s", userFrom, exception, specs, userComment)
	message.Files = []*discordgo.File{file}

	msg, err := discord.ChannelMessageSendComplex(config.Discord.Output, message)
	if err != nil {
		return err
	}

	msgs[msg.ID] = message.Content
	return err
}

func emailRoutine(kill chan bool) {
	for {
		readEmail()

		for i := 0; i < (config.Email.Every / 10); i++ {
			select {
			case <-kill:
				return
			default:
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func readEmail() {
	conn, err := pop3.DialTLS(config.Email.Host)
	if err != nil {
		fmt.Printf("Could not dial %s: %v\n", config.Email.Host, err)
		return
	}

	defer conn.Quit()

	if err = conn.Auth(config.Email.User, config.Email.Pass); err != nil {
		fmt.Printf("Could not auth %s: %v\n", config.Email.User, err)
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

			exception, specs, err := parseCrash(zipReader)
			if err != nil {
				fmt.Printf("Could not parse crash in mail %d\n", msgid)
				continue
			}

			postMessage(zipByteReader, mail.TextBody, "`"+mail.From[0].Address+"`", exception, specs)

			if err := conn.Dele(msgid); err != nil {
				fmt.Printf("Could not delete mail %d\n", msgid)
				continue
			}
		}
	}
}
