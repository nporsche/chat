package main

const (
	bufferSize = 1024
)

type Channel struct {
	Name          string
	People        []*Client
	messageBuffer chan string
}

func NewChannel(name string) *Channel {
	this := new(Channel)
	this.Name = name
	this.People = []*Client{}
	this.messageBuffer = make(chan string, bufferSize)
	go this.doDistributeMessage()
	return this
}

func (this *Channel) Join(c *Client) {
	existed := false
	for _, ele := range this.People {
		if ele == c {
			existed = true
			break
		}
	}
	if !existed {
		this.People = append(this.People, c)
	}
}

func (this *Channel) Quit(c *Client) {
	for i, ele := range this.People {
		if ele == c {
			this.People[i], this.People[len(this.People)-1] = this.People[len(this.People)-1], (this.People)[i]
			this.People = this.People[:len(this.People)-1]
			break
		}
	}
}

func (this *Channel) PostMessage(message string) {
	this.messageBuffer <- message
}

func (this *Channel) doDistributeMessage() {
	for {
		msg := <-this.messageBuffer
		for _, person := range this.People {
			person.WriteLine(msg)
		}
	}
}
