package main

import (
	"fmt"
	"time"
	"log"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"github.com/bwmarrin/discordgo"
	"bufio"
)

var Token string

type command struct {
	cmd string
	param string
	user string
	roles []string
	guild string
}

var channels map[string]string
var pronouns map[string]string

func init() {
	fmt.Println("Initializing...")
	getCreds()
	channels = make(map[string]string)
	pronouns = make(map[string]string)
	pronouns["he/him"] = "641681631605817395"
	pronouns["she/her"] = "641681795603234861"
	pronouns["he/they"] = "641681957520015363"
	pronouns["she/they"] = "641681913093947413"
	pronouns["they/them"] = "641681835180425216"
	pronouns["she/he/they"] = "663915236742529024"
	pronouns["any"] = "663882262680436736"
	channels["pronouns"] =  "641684780408242198"
}

func  main() {
	dg, err := discordgo.New("Bot " + Token)
	fmt.Println("Establishing Discord session with token...")

	if err != nil {
		fmt.Println("Failed to establish Discord session: ", err)
		return
	}

	// create a handler for new messages
	dg.AddHandler(messageCreate)

	// open a websocket connection to discord
	err = dg.Open()
	fmt.Println("Opening websocket connection to Discord...")

	if err != nil {
		fmt.Println("Error opening connection to Discord:", err)
		fmt.Println("Token used:", Token)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	var changelog = getChangelog()
	if(!strings.Contains(changelog[0], "null")) {
		dg.ChannelMessageSend(channels["pronouns"], "Heyo! I'm back from the maintenance man. Here's what changed:")
		for _, eachline := range changelog {
			dg.ChannelMessageSend(channels["pronouns"], eachline)
		}
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<- sc

	dg.Close()
}

// message handler
func messageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {

	if msg.Author.ID == s.State.User.ID {
		return
	}

	if msg.ChannelID == channels["pronouns"] {
		msgCmd := command {
			cmd: "",
			param: "",
			user: msg.Author.ID,
			roles: msg.Member.Roles,
			guild: msg.GuildID,
		}

		reply := "All set! I've got you down as "
		list := "Here are available roles:\nhe/him\nshe/her\nthey/them\nshe/they\nhe/they\nany\nhe/she/they"
		msg.Content = strings.ToLower(msg.Content)
		if (strings.Contains(msg.Content, "--help")) {
			s.ChannelMessageSend(msg.ChannelID, "Hey y'all! Jacob's Mechanical Overlord here to help!\nHere are some pronoun commands:\n-To set, just type the pronouns you want(i.e. 'they/them')\n-To remove pronouns, type remove before the ones that you want removed(i.e. 'remove he/him')\n-To change pronouns, type change before the ones you're changing to(i.e. 'change she/her')")
			s.ChannelMessageSend(msg.ChannelID, list)
		} else if (strings.Contains(msg.Content, "change")) {
			msgCmd.cmd = "change"
			msgCmd.param = strings.ReplaceAll(msg.Content, "change", "")
			msgCmd.param = strings.ToLower(msgCmd.param)
		} else if (strings.Contains(msg.Content, "remove")) {
			msgCmd.cmd = "remove"
			msgCmd.param = strings.ReplaceAll(msg.Content, "remove", "")
			msgCmd.param = strings.ToLower(msgCmd.param)
		} else {
			msgCmd.cmd = "set"
			msgCmd.param = strings.ToLower(msg.Content)
		}

		if msgCmd.cmd == "set" {
			err := s.GuildMemberRoleAdd(msgCmd.guild, msgCmd.user, pronouns[msgCmd.param])
			if err != nil {
				fmt.Println("Error setting role:", err)
				s.ChannelMessageSend(msg.ChannelID, "Hmm, something seems to be wrong with that command.")
				return
			}
			s.ChannelMessageSend(msg.ChannelID, (reply + strings.TrimSpace(msgCmd.param)))
		} else if msgCmd.cmd == "remove" {
			err := s.GuildMemberRoleRemove(msgCmd.guild, msgCmd.user, pronouns[strings.TrimSpace(msgCmd.param)])
                        if err != nil {
                                fmt.Println("Error removing role:", err)
				fmt.Printf("Using:\nGuild: %s\nUser: %s\nRole:%s\n", msgCmd.guild, msgCmd.user, pronouns[msgCmd.param])
				s.ChannelMessageSend(msg.ChannelID, "Hmm, something seems to be wrong with that command.")
				return
                        }
			s.ChannelMessageSend(msg.ChannelID, ("No problemo, I'll remove " + strings.TrimSpace(msgCmd.param)) + " from your roles.")
		} else if msgCmd.cmd == "change" {
			for key, value := range pronouns {
				fmt.Println("Key:", key)
				for i := 0; i < len(msgCmd.roles); i++ {
					if msgCmd.roles[i] == value {
						err := s.GuildMemberRoleRemove(msgCmd.guild, msgCmd.user, strings.TrimSpace(value))
						if err != nil {
							fmt.Println("Error changing roles:", err)
							s.ChannelMessageSend(msg.ChannelID, "Hmm, something seems to be wrong with that command.")
							return
						}
					}
				}
			}
			err := s.GuildMemberRoleAdd(msgCmd.guild, msgCmd.user, pronouns[strings.TrimSpace(msgCmd.param)])
                        if err != nil {
                                fmt.Println("Error changing roles:", err)
				s.ChannelMessageSend(msg.ChannelID, "Hmm, something seems to be wrong with that command.")
				return
                        }
			s.ChannelMessageSend(msg.ChannelID, (reply + strings.TrimSpace(msgCmd.param)))
		}

		fmt.Printf("Command: %s \n Parameter: %s \n User: %s \n Role: %s \n", msgCmd.cmd, msgCmd.param, msgCmd.user, msgCmd.roles)
	}
}

func getCreds() {
	file, err := os.Open("creds.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	data := fmt.Sprintf("%s", b)
	Token = strings.TrimSpace(data)
}

func getChangelog() []string {
	file, err := os.Open("changelog.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	loc, _ := time.LoadLocation("America/Chicago")

	dt := time.Now().In(loc)

	s := bufio.NewScanner(file)
	s.Split(bufio.ScanLines)
	var lines []string
	var appendLog = false
	var announce = false
	var aC = 0
	for s.Scan() {
		if(strings.Contains(s.Text(), dt.Format("01-02-2006"))) {
			appendLog = true
			announce = true
		} else if(strings.Contains(s.Text(), "*")) {
			appendLog = false
			fmt.Println(dt.Format("01-02-2006"))
		}

		if(appendLog && (aC != 0)) {
			lines = append(lines, s.Text())
			fmt.Println("appended " + s.Text())
		}
		aC += 1
	}
	if(!announce) {
		lines = append(lines, "null")
		fmt.Println("!announce")
	}
	return lines
}
