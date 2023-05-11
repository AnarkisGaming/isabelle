package discord

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/AnarkisGaming/isabelle/config"
	"github.com/AnarkisGaming/isabelle/crash"
	"github.com/AnarkisGaming/isabelle/github"
	"github.com/bwmarrin/discordgo"
)

var (
	discord *discordgo.Session
	msgs    map[string]string
)

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
		reply(s, m, fmt.Sprintf(config.Config.Messages.Help, m.Author.Mention()))
		return
	}

	if !strings.HasSuffix(m.Attachments[0].Filename, ".zip") {
		reply(s, m, config.Config.Messages.InvalidFile)
		return
	}

	fmt.Printf("DM received from %s#%s (%s) containing %s\n", m.Author.Username, m.Author.Discriminator, m.Author.ID, m.Attachments[0].URL)

	resp, err := http.Get(m.Attachments[0].URL)
	if err != nil {
		reply(s, m, config.Config.Messages.BadDownload)
		fmt.Printf("Could not obtain uploaded file %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	defer resp.Body.Close()

	// zips cannot be streamed
	zipBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reply(s, m, config.Config.Messages.BadDownload)
		fmt.Printf("Could not read uploaded file stream from %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	byteReader := bytes.NewReader(zipBytes)

	zipReader, err := zip.NewReader(byteReader, int64(byteReader.Len()))
	if err != nil {
		reply(s, m, config.Config.Messages.BadZIP)
		fmt.Printf("Could not obtain valid zip from %s: %v\n", m.Attachments[0].URL, err)
		return
	}

	exception, specs, err := crash.Parse(zipReader)
	if err != nil {
		reply(s, m, config.Config.Messages.BadZIP)
		return
	}

	if len(exception) == 0 || len(specs) == 0 {
		reply(s, m, config.Config.Messages.BadZIP)
		fmt.Printf("%s was missing some files\n", m.Attachments[0].URL)
		return
	}

	PostMessage(byteReader, m.Content, m.Author.Mention(), exception, specs)

	reply(s, m, config.Config.Messages.Success)
}

func handleReact(s *discordgo.Session, react *discordgo.MessageReactionAdd) {
	if react.Emoji.Name != config.Config.GitHub.React {
		return
	}

	msg, ok := msgs[react.MessageID]
	if !ok {
		return
	}

	issueURL, err := github.CreateIssue(
		"Crash report: "+strings.Split(msg, "\n")[2],
		"Issue reported via Isabelle: https://discord.com/channels/"+react.MessageReaction.GuildID+"/"+config.Config.Discord.Output+"/"+react.MessageID+"/\n\n"+strings.Join(strings.Split(msg, "\n")[1:], "\n"))

	if err != nil {
		s.ChannelMessageSend(react.ChannelID, fmt.Sprintf("Could not open issue: `%v`", err))
		return
	}

	_, err = s.ChannelMessageSend(react.ChannelID, fmt.Sprintf("<@"+react.UserID+">: Opened issue <%s>", issueURL))
	if err != nil {
		fmt.Printf("Could not send message: %v\n", err)
		return
	}
}

// PostMessage posts a message to the configured Discord guild with crash info
func PostMessage(byteReader *bytes.Reader, userComment string, userFrom string, exception string, specs string) (err error) {
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

	msg, err := discord.ChannelMessageSendComplex(config.Config.Discord.Output, message)
	if err != nil {
		return err
	}

	msgs[msg.ID] = message.Content
	return err
}

// Connect connects to Discord and throws if the connection is unsuccessful
func Connect() error {
	var err error

	discord, err = discordgo.New("Bot " + config.Config.Discord.Token)
	if err != nil {
		return err
	}

	discord.AddHandler(handleDM)
	discord.AddHandler(handleReact)

	if err = discord.Open(); err != nil {
		return err
	}

	msgs = make(map[string]string)

	fmt.Printf("Successfully connected to Discord as %s\n", discord.State.User.String())

	return nil
}

// Close closes the Discord instance
func Close() error {
	return discord.Close()
}
