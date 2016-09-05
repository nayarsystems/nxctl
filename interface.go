package main

import "github.com/nayarsystems/kingpin"

const (
	DEFAULT_USER    = "test"
	DEFAULT_PASS    = "test"
	DEFAULT_SERVER  = "127.0.0.1:1717"
	DEFAULT_TIMEOUT = 60
)

var (
	app       = kingpin.New("cli", "Nexus command line interface")
	serverIP  = app.Flag("server", "Server address.").Short('s').String()
	timeout   = app.Flag("timeout", "Execution timeout").Short('t').Int()
	user      = app.Flag("user", "Nexus username").Short('u').String()
	pass      = app.Flag("pass", "Nexus password").Short('p').String()
	config    = app.Flag("config", "Config filename").Short('c').String()
	ignoreapi = app.Flag("ignoreapi", "Ignore API version check").Bool()

	///

	shell = app.Command("shell", "Interactive shell")

	///

	login     = app.Command("login", "Tests to login with an username/password and exits")
	loginName = login.Arg("username", "username").Required().String()
	loginPass = login.Arg("password", "password").Required().String()

	///

	push       = app.Command("push", "Execute a task.push rpc call on Nexus")
	pushMethod = push.Arg("method", "Method to call").Required().String()
	pushParams = push.Arg("params", "parameters").StringMap()

	pull       = app.Command("pull", "Execute a task.pull rpc call on Nexus")
	pullMethod = pull.Arg("prefix", "Method to call").Required().String()

	taskList       = app.Command("list", "Show push/pulls happening on a prefix")
	taskListPrefix = taskList.Arg("prefix", "prefix").Default("").String()
	taskListLimit  = taskList.Flag("limit", "Limit the number of tasks returned").Default("100").Int()
	taskListSkip   = taskList.Flag("skip", "Skip a number of tasks before applying the limit").Default("0").Int()

	///

	pipeCmd = app.Command("pipe", "Pipe tasks")

	pipeRead = pipeCmd.Command("read", "Create and read from a pipe. It will be destroyed on exit")

	pipeWrite     = pipeCmd.Command("open", "Open a pipe and send data")
	pipeWriteId   = pipeWrite.Arg("pipeId", "ID of the pipe to write to").Required().String()
	pipeWriteData = pipeWrite.Arg("data", "Data to write to the pipe").Required().Strings()

	///

	userCmd = app.Command("user", "user management")

	userCreate     = userCmd.Command("create", "Create a new user")
	userCreateName = userCreate.Arg("username", "username").Required().String()
	userCreatePass = userCreate.Arg("password", "password").Required().String()

	userDelete     = userCmd.Command("delete", "Delete an user")
	userDeleteName = userDelete.Arg("username", "username").Required().String()

	userPass     = userCmd.Command("passwd", "Change an user password")
	userPassName = userPass.Arg("username", "username").Required().String()
	userPassPass = userPass.Arg("password", "password").Required().String()

	userList       = userCmd.Command("list", "List users on a prefix")
	userListPrefix = userList.Arg("prefix", "prefix").Default("").String()
	userListLimit  = userList.Flag("limit", "Limit the number of users returned").Default("100").Int()
	userListSkip   = userList.Flag("skip", "Skip a number of elements before applying the limit").Default("0").Int()

	userKick       = userCmd.Command("kick", "Kick users on a prefix")
	userKickPrefix = userKick.Arg("prefix", "prefix").Required().String()

	userReload       = userCmd.Command("reload", "Reloads users on a prefix")
	userReloadPrefix = userReload.Arg("prefix", "prefix").Required().String()

	userMaxSessions     = userCmd.Command("max-sessions", "Sets the maximum number of sessions for an user")
	userMaxSessionsUser = userMaxSessions.Arg("username", "username").Required().String()
	userMaxSessionsN    = userMaxSessions.Arg("max", "max").Required().Int()

	///

	sessionsCmd = app.Command("sessions", "Show sessions info")

	sessionsList       = sessionsCmd.Command("list", "List active sessions")
	sessionsListPrefix = sessionsList.Arg("prefix", "User prefix").Default("").String()
	sessionsListLimit  = sessionsList.Flag("limit", "Limit the number of sessions returned").Default("100").Int()
	sessionsListSkip   = sessionsList.Flag("skip", "Skip a number of elements before applying the limit").Default("0").Int()

	sessionsKick     = sessionsCmd.Command("kick", "Kick any active connection with matching prefix")
	sessionsKickConn = sessionsKick.Arg("connId", "connId prefix").Required().String()

	sessionsReload     = sessionsCmd.Command("reload", "Reload any active connection with matching prefix")
	sessionsReloadConn = sessionsReload.Arg("connId", "connId prefix").Required().String()

	///

	nodesCmd      = app.Command("nodes", "Show nodes info")
	nodesCmdLimit = nodesCmd.Flag("limit", "Limit the number of nodes returned").Default("100").Int()
	nodesCmdSkip  = nodesCmd.Flag("skip", "Skip a number of elements before applying the limit").Default("0").Int()

	///

	tagsCmd = app.Command("tags", "tags management")

	tagsSet       = tagsCmd.Command("set", "Set tags for an user on a prefix. Tags is a map like 'tag:value tag2:value2'")
	tagsSetUser   = tagsSet.Arg("user", "user").Required().String()
	tagsSetPrefix = tagsSet.Arg("prefix", "prefix").Required().String()
	tagsSetTags   = tagsSet.Arg("tags", "tag:value").StringMapIface()

	tagsSetJ         = tagsCmd.Command("setj", "Set tags for an user on a prefix. Tags is a json dict like: { 'tag': value }")
	tagsSetJUser     = tagsSetJ.Arg("user", "user").Required().String()
	tagsSetJPrefix   = tagsSetJ.Arg("prefix", "prefix").Required().String()
	tagsSetJTagsJson = tagsSetJ.Arg("json {tag:value,...}", "{'@task.push': true}").Required().String()

	tagsDel       = tagsCmd.Command("del", "Delete tags for an user on a prefix. Tags is a list of space separated strings")
	tagsDelUser   = tagsDel.Arg("user", "user").Required().String()
	tagsDelPrefix = tagsDel.Arg("prefix", "prefix").Required().String()
	tagsDelTags   = tagsDel.Arg("tags", "tag1 tag2 tag3").Required().Strings()

	//

	templateCmd = app.Command("template", "template management")

	templateAdd         = templateCmd.Command("add", "Add a template to the user")
	templateAddUser     = templateAdd.Arg("user", "user").Required().String()
	templateAddTemplate = templateAdd.Arg("template", "template").Required().String()

	templateDel         = templateCmd.Command("del", "Removes a template from the user")
	templateDelUser     = templateDel.Arg("user", "user").Required().String()
	templateDelTemplate = templateDel.Arg("template", "template").Required().String()

	//

	whitelistCmd = app.Command("whitelist", "Whitelist management")

	whitelistAdd     = whitelistCmd.Command("add", "Whitelist an address for the user")
	whitelistAddUser = whitelistAdd.Arg("user", "user").Required().String()
	whitelistAddIP   = whitelistAdd.Arg("ip", "ip regex").Required().String()

	whitelistDel     = whitelistCmd.Command("del", "Removes a whitelisted address from the user")
	whitelistDelUser = whitelistDel.Arg("user", "user").Required().String()
	whitelistDelIP   = whitelistDel.Arg("ip", "ip regex").Required().String()

	//

	blacklistCmd = app.Command("blacklist", "Blacklist management")

	blacklistAdd     = blacklistCmd.Command("add", "Blacklist an address for the user")
	blacklistAddUser = blacklistAdd.Arg("user", "user").Required().String()
	blacklistAddIP   = blacklistAdd.Arg("ip", "ip regex").Required().String()

	blacklistDel     = blacklistCmd.Command("del", "Removes a blacklist from the user")
	blacklistDelUser = blacklistDel.Arg("user", "user").Required().String()
	blacklistDelIP   = blacklistDel.Arg("ip", "ip regex").Required().String()

	//

	chanCmd = app.Command("topic", "Topics management")

	chanSub     = chanCmd.Command("sub", "Subscribe a pipe to a topic")
	chanSubPipe = chanSub.Arg("pipe", "pipe id to subscribe").Required().String()
	chanSubChan = chanSub.Arg("topic", "Topic to subscribe to").Required().String()

	chanUnsub     = chanCmd.Command("unsub", "Unsubscribe a pipe from a topic")
	chanUnsubPipe = chanUnsub.Arg("pipe", "pipe id to subscribe").Required().String()
	chanUnsubChan = chanUnsub.Arg("topic", "Topic to subscribe to").Required().String()

	chanPub     = chanCmd.Command("pub", "Publish a message to a topic")
	chanPubChan = chanPub.Arg("topic", "Topic to subscribe to").Required().String()
	chanPubMsg  = chanPub.Arg("data", "Data to send").Required().Strings()
)
