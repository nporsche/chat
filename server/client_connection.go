package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	LINE_BUF_SIZE = 1024
)

const (
	LOGIN_PREFIX = `\user `
	JOIN_PREFIX  = `\join `
	HELP_PREFIX  = `\help`
)

type Client struct {
	conn      net.Conn
	ioBuf     *bufio.ReadWriter
	msgBuffer chan string
	curCh     *Channel
	chMgr     *ChannelManager

	userName   string
	email      string
	serverName string
	realName   string
}

func NewClient(conn net.Conn, chMgr *ChannelManager) *Client {
	this := new(Client)
	this.conn = conn
	this.msgBuffer = make(chan string, bufferSize)
	this.chMgr = chMgr
	this.ioBuf = bufio.NewReadWriter(bufio.NewReaderSize(conn, LINE_BUF_SIZE), bufio.NewWriter(conn))
	go this.doWrite()

	return this
}

/*
protocol description:
1. \join <channel>
join the channel. If already in a channel, it will first quit then join the new one.

2. \user <username> <email> <servername> <realname>
login into the server
*/
func (this *Client) MainLoop() {
	line, err := this.ReadLine()
	if err != nil {
		log.Printf("Read Login error=[%v]\n", err)
		this.Close()
		return
	}

	if n, err := fmt.Sscanf(line, "\\user %s %s %s %s", &this.userName, &this.email, &this.serverName, &this.realName); err != nil || n != 4 {
		log.Println("invalid Login format")
		this.Close()
		return
	}

	for {
		//retrieve full line
		line, err := this.ReadLine()
		if err != nil {
			log.Println("ReadLine error")
			this.Close()
			break
		}

		log.Printf("recv content=[%s] from=[%s]\n", line, this.conn.RemoteAddr())

		//this line is command
		if this.HandleCommand(line) {
			continue
		}

		//this line is normal message
		this.PostChanMessage(line)
	}
}

//Post message to current channel
func (this *Client) PostChanMessage(message string) {
	if this.curCh != nil {
		this.curCh.PostMessage(message)
	} else {
		this.WriteLine("You should join a channel first")
	}
}

func (this *Client) WriteLine(line string) {
	this.msgBuffer <- line
}

func (this *Client) ReadLine() (string, error) {
	buffer := bytes.NewBufferString("")
	for {
		lineBytes, prefix, err := this.ioBuf.ReadLine()
		if err != nil {
			return "", err
		}
		buffer.Write(lineBytes)
		if !prefix {
			break
		}
	}
	return buffer.String(), nil
}

func (this *Client) Close() {
	if this.curCh != nil {
		this.curCh.Quit(this)
	}
	this.conn.Close()
	log.Printf("%s has been quit\n", this.conn.RemoteAddr())
}

func (this *Client) HandleCommand(cmd string) bool {
	if cmd[0] != '\\' {
		return false
	}

	if strings.HasPrefix(cmd, JOIN_PREFIX) {
		chName := cmd[len(JOIN_PREFIX):]
		ch, err := this.chMgr.Channel(chName)
		if err != nil {
			this.WriteLine(fmt.Sprintf("%s channel does not exist", chName))
			return true
		}
		if this.curCh != nil {
			this.PostChanMessage(fmt.Sprintf("%s has been left", this.userName))
			this.curCh.Quit(this)
		}
		ch.Join(this)
		this.curCh = ch
		this.PostChanMessage(fmt.Sprintf("%s has been joined", this.userName))
	} else if strings.HasPrefix(cmd, HELP_PREFIX) {
		this.WriteLine("commands:\r\n\\join <channel>  ***join the channel. If already in a channel, it will first quit then join the new one.\r\n\\user <username> <email> <servername> <realname>  ***login into the server\r\n\\help  ***get supported commands")
	} else {
		this.WriteLine("Invalid Command")
	}

	return true
}

func (this *Client) doWrite() {
	for {
		msg := <-this.msgBuffer

		_, err := this.ioBuf.WriteString(msg)
		if err != nil {
			break
		}
		_, err = this.ioBuf.WriteString("\r\n")
		if err != nil {
			break
		}
		err = this.ioBuf.Flush()
		if err != nil {
			break
		}
	}
}
