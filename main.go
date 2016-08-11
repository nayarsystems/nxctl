package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaracil/ei"
	nxcli "github.com/jaracil/nxcli"
	nexus "github.com/jaracil/nxcli/nxcore"
	"github.com/nayarsystems/kingpin"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

func main() {

	// Enable -h as HelpFlag
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	parsed := kingpin.MustParse(app.Parse(os.Args[1:]))

	viper.SetConfigName(".nxctl")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	if *user == "" {
		if viper.IsSet("user") {
			*user = viper.GetString("user")
		} else {
			*user = DEFAULT_USER
		}
	}
	if *pass == "" {
		if viper.IsSet("pass") {
			*pass = viper.GetString("pass")
		} else {
			*pass = DEFAULT_PASS
		}
	}

	if *serverIP == "" {
		if viper.IsSet("server") {
			*serverIP = viper.GetString("server")
		} else {
			*serverIP = DEFAULT_SERVER
		}
	}

	if *timeout == 0 {
		if viper.IsSet("timeout") {
			*timeout = viper.GetInt("timeout")
		} else {
			*timeout = DEFAULT_TIMEOUT
		}
	}

	if nc, err := nxcli.Dial(*serverIP, nil); err == nil {
		log.Println("Connected to", *serverIP)
		exec(nc, parsed)
	} else {
		log.Printf("Cannot connect to %s: %s\n", *serverIP, err)
	}
}

func exec(nc *nexus.NexusConn, parsed string) {
	if parsed == login.FullCommand() {
		if _, err := nc.Login(*loginName, *loginPass); err != nil {
			log.Println("Couldn't login:", err)
			return
		} else {
			log.Println("Logged as", *loginName)
			user = loginName
		}
		return
	}

	if res, err := nc.Login(*user, *pass); err != nil {
		log.Println("Couldn't login:", err)
		return
	} else {
		if ei.N(res).M("ok").BoolZ() {
			log.Println("Logged as", ei.N(res).M("user").StringZ())
		} else {
			log.Println("Unexpected reply:", res)
			return
		}
	}

	execCmd(nc, parsed)
}

func execCmd(nc *nexus.NexusConn, parsed string) {
	switch parsed {
	case push.FullCommand():
		if ret, err := nc.TaskPush(*pushMethod, *pushParams, time.Second*time.Duration(*timeout)); err != nil {
			log.Println("Error:", err)
			return
		} else {
			b, _ := json.MarshalIndent(ret, "", "  ")
			log.Println("Result:")
			if s, err := strconv.Unquote(string(b)); err == nil {
				fmt.Println(s)
			} else {
				fmt.Println(string(b))
			}
		}

	case pull.FullCommand():
		log.Println("Pulling", *pullMethod)
		ret, err := nc.TaskPull(*pullMethod, time.Second*time.Duration(*timeout))
		if err != nil {
			log.Println("Error:", err)
			return
		} else {
			b, _ := json.MarshalIndent(ret, "", "  ")
			fmt.Println(string(b))
		}

		fmt.Printf("[R]esult or [E]rror? ")

		stdin := bufio.NewScanner(os.Stdin)

		if stdin.Scan() && strings.HasPrefix(strings.ToLower(stdin.Text()), "e") {
			fmt.Printf("Code: ")
			stdin.Scan()
			code, _ := strconv.Atoi(stdin.Text())

			fmt.Printf("Message: ")
			stdin.Scan()
			msg := stdin.Text()

			fmt.Printf("Data: ")
			stdin.Scan()
			data := stdin.Text()

			ret.SendError(code, msg, data)

		} else {
			fmt.Printf("Result: ")
			if stdin.Scan() {
				ret.SendResult(stdin.Text())
			} else {
				ret.SendResult("dummy response")
			}
		}

	case taskList.FullCommand():
		if res, err := nc.TaskList(*taskListPrefix, *taskListLimit, *taskListSkip); err != nil {
			log.Println(err)
			return
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Session", "ID", "Path", "Method", "Params", "User", "State", "Worker"})
			table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})

			for _, task := range res {
				table.Append([]string{task.Id[:16], task.Id[16:], task.Path, task.Method, fmt.Sprintf("%v", task.Params), task.User, task.Stat, task.Tses})
			}
			table.Render()
		}

	case pipeWrite.FullCommand():
		// Clean afterwards in case we are looping on shell mode
		defer func() { *pipeWriteData = []string{} }()

		if pipe, err := nc.PipeOpen(*pipeWriteId); err != nil {
			log.Println(err)
			return
		} else {

			if _, err := pipe.Write(*pipeWriteData); err != nil {
				log.Println(err)
				return
			} else {
				log.Println("Sent!")
			}
		}

	case pipeRead.FullCommand():
		popts := nexus.PipeOpts{Length: 100}

		if pipe, err := nc.PipeCreate(&popts); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("Pipe created:", pipe.Id())
			for {
				if pdata, err := pipe.Read(10, time.Second*time.Duration(*timeout)); err != nil {
					log.Println(err)
					time.Sleep(time.Second)
				} else {
					for _, msg := range pdata.Msgs {
						log.Println("Got:", msg.Msg, msg.Count)
					}
					fmt.Printf("There are %d messages left in the pipe and %d drops\n", pdata.Waiting, pdata.Drops)
				}
			}
		}

	case userCreate.FullCommand():
		log.Printf("Creating user \"%s\" with password \"%s\"", *userCreateName, *userCreatePass)
		if _, err := nc.UserCreate(*userCreateName, *userCreatePass); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case userDelete.FullCommand():
		log.Printf("Deleting user \"%s\"", *userDeleteName)

		if _, err := nc.UserDelete(*userDeleteName); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case userList.FullCommand():
		log.Printf("Listing users on \"%s\"", *userListPrefix)

		if res, err := nc.UserList(*userListPrefix, *userListLimit, *userListSkip); err != nil {
			log.Println(err)
			return
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"User", "Templates", "Whitelist", "Blacklist", "Max Sessions", "Prefix", "Tags"})
			table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
			table.SetAlignment(tablewriter.ALIGN_CENTER)
			table.SetRowLine(true)
			//table.SetRowSeparator(".")

			for _, user := range res {
				lines := 0
				for prefix, tags := range user.Tags {
					for tag, val := range tags {
						if lines == 0 {
							table.Append([]string{user.User, fmt.Sprintf("%v", user.Templates), fmt.Sprintf("%v", user.Whitelist), fmt.Sprintf("%v", user.Blacklist), fmt.Sprintf("%d", user.MaxSessions), prefix, fmt.Sprintf("%s: %v", tag, val)})
						} else {
							table.Append([]string{"", "", "", "", "", prefix, fmt.Sprintf("%s: %v", tag, val)})
						}
						lines++
					}
				}

				if lines == 0 {
					table.Append([]string{user.User, fmt.Sprintf("%v", user.Templates), fmt.Sprintf("%v", user.Whitelist), fmt.Sprintf("%v", user.Blacklist), fmt.Sprintf("%d", user.MaxSessions)})
				}
			}

			table.Render() // Send output
			fmt.Println()
		}

	case userPass.FullCommand():
		if _, err := nc.UserSetPass(*userPassName, *userPassPass); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case userMaxSessions.FullCommand():
		if _, err := nc.UserSetMaxSessions(*userMaxSessionsUser, *userMaxSessionsN); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case userKick.FullCommand():
		log.Printf("Kicking users on \"%s\"", *userKickPrefix)

		if res, err := nc.SessionList(*userKickPrefix, -1, -1); err != nil {
			log.Println(err)
			return
		} else {
			for _, session := range res {
				log.Printf("\tUser: [%s] - %d sessions", session.User, session.N)
				for _, ses := range session.Sessions {
					if kicked, err := nc.SessionKick(ses.Id); err == nil && ei.N(kicked).M("kicked").IntZ() == 1 {
						log.Printf("\t\tID: %s has been kicked", ses.Id)
					}
				}
			}
		}

	case userReload.FullCommand():
		log.Printf("Reloading users on \"%s\"", *userReloadPrefix)

		if res, err := nc.SessionList(*userReloadPrefix, -1, -1); err != nil {
			log.Println(err)
			return
		} else {
			for _, session := range res {
				log.Printf("\tUser: [%s] - %d sessions", session.User, session.N)
				for _, ses := range session.Sessions {
					if reloaded, err := nc.SessionReload(ses.Id); err == nil && ei.N(reloaded).M("reloaded").IntZ() == 1 {
						log.Printf("\t\tID: %s has been reloaded", ses.Id)
					}
				}
			}
		}

	case sessionsList.FullCommand():
		if res, err := nc.SessionList(*sessionsListPrefix, *sessionsListLimit, *sessionsListSkip); err != nil {
			log.Println(err)
			return
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Node", "User", "Protocol", "Remote Addr", "Since"})
			table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
			n := 0
			for _, session := range res {
				for _, ses := range session.Sessions {
					n++
					table.Append([]string{ses.Id, ses.NodeId, session.User, ses.Protocol, ses.RemoteAddress, ses.CreationTime.Format("Mon Jan _2 15:04:05 2006")})
				}
				table.Append([]string{"Sessions:", fmt.Sprintf("%d", session.N), "", "", "", ""})
				table.Append([]string{"", "", "", "", "", ""})

			}
			table.Append([]string{"Total Sessions:", fmt.Sprintf("%d", n), "", "", "", ""})

			table.Render() // Send output
			fmt.Println()

		}

	case sessionsKick.FullCommand():
		if res, err := nc.SessionKick(*sessionsKickConn); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("Sessions kicked:", ei.N(res).M("kicked").IntZ())
		}

	case sessionsReload.FullCommand():
		if res, err := nc.SessionReload(*sessionsReloadConn); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("Sessions reloaded:", ei.N(res).M("reloaded").IntZ())
		}

	case nodesCmd.FullCommand():
		if res, err := nc.NodeList(*nodesCmdLimit, *nodesCmdSkip); err != nil {
			log.Println(err)
			return
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Node", "Clients", "Load"})
			table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
			table.SetAlignment(tablewriter.ALIGN_CENTER)

			n := 0
			for _, node := range res {
				n += node.Clients
				table.Append([]string{node.NodeId, fmt.Sprintf("%d", node.Clients), fmt.Sprintf("%0.2f / %0.2f / %0.2f", node.Load["Load1"], node.Load["Load5"], node.Load["Load15"])})
			}

			table.SetFooter([]string{fmt.Sprintf("%d", len(res)), fmt.Sprintf("%d", n), ""})
			table.Render() // Send output
		}

	case tagsSet.FullCommand():
		// Clean afterwards in case we are looping on shell mode
		defer func() { *tagsSetTags = make(map[string]interface{}) }()

		var tags map[string]interface{}
		if b, err := json.Marshal(*tagsSetTags); err == nil {
			if json.Unmarshal(b, &tags) != nil {
				log.Println("Error parsing tags")
				return
			}
		}

		log.Printf("Setting tags: %v on %s@%s", tags, *tagsSetUser, *tagsSetPrefix)
		if _, err := nc.UserSetTags(*tagsSetUser, *tagsSetPrefix, tags); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case tagsSetJ.FullCommand():
		// Clean afterwards in case we are looping on shell mode

		var tags map[string]interface{}
		if json.Unmarshal([]byte(*tagsSetJTagsJson), &tags) != nil {
			log.Println("Error parsing tags json:", *tagsSetJTagsJson)
			return
		}

		log.Printf("Setting tags: %v on %s@%s", tags, *tagsSetJUser, *tagsSetJPrefix)
		if _, err := nc.UserSetTags(*tagsSetJUser, *tagsSetJPrefix, tags); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case tagsDel.FullCommand():
		// Clean afterwards in case we are looping on shell mode
		defer func() { *tagsDelTags = []string{} }()

		if _, err := nc.UserDelTags(*tagsDelUser, *tagsDelPrefix, *tagsDelTags); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case templateAdd.FullCommand():
		if _, err := nc.UserAddTemplate(*templateAddUser, *templateAddTemplate); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case templateDel.FullCommand():
		if _, err := nc.UserDelTemplate(*templateDelUser, *templateDelTemplate); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case whitelistAdd.FullCommand():
		if _, err := nc.UserAddWhitelist(*whitelistAddUser, *whitelistAddIP); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case whitelistDel.FullCommand():
		if _, err := nc.UserDelWhitelist(*whitelistDelUser, *whitelistDelIP); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case blacklistAdd.FullCommand():
		if _, err := nc.UserAddBlacklist(*blacklistAddUser, *blacklistAddIP); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case blacklistDel.FullCommand():
		if _, err := nc.UserDelBlacklist(*blacklistDelUser, *blacklistDelIP); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("OK")
		}

	case shell.FullCommand():

		args := os.Args[1:]
		for k, v := range args {
			if v == shell.FullCommand() {
				args = append(args[:k], args[k+1:]...)
			}
		}

		s := bufio.NewScanner(os.Stdin)
		fmt.Printf("%s@%s >> ", *user, *serverIP)
		for s.Scan() {
			cmd, err := app.Parse(append(args, strings.Split(s.Text(), " ")...))
			if err == nil {
				if cmd != shell.FullCommand() {
					parsed := kingpin.MustParse(cmd, err)
					execCmd(nc, parsed)
				}
			} else {
				log.Println(err)
			}
			fmt.Printf("%s@%s >> ", *user, *serverIP)
		}

		if err := s.Err(); err != nil {
			log.Fatalln("reading standard input:", err)
		}

	case chanSub.FullCommand():
		if pipe, err := nc.PipeOpen(*chanSubPipe); err != nil {
			log.Println(err)
			return
		} else {
			if _, err := nc.TopicSubscribe(pipe, *chanSubChan); err != nil {
				log.Println(err)
				return
			} else {
				log.Println("OK")
			}
		}

	case chanUnsub.FullCommand():
		if pipe, err := nc.PipeOpen(*chanSubPipe); err != nil {
			log.Println(err)
			return
		} else {
			if _, err := nc.TopicUnsubscribe(pipe, *chanUnsubChan); err != nil {
				log.Println(err)
				return
			} else {
				log.Println("OK")
			}
		}

	case chanPub.FullCommand():
		// Clean afterwards in case we are looping on shell mode
		defer func() { *chanPubMsg = []string{} }()

		if res, err := nc.TopicPublish(*chanPubChan, *chanPubMsg); err != nil {
			log.Println(err)
			return
		} else {
			log.Println("Result:", res)

		}
	}
}
