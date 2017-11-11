package pageChannel

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
)

const (
	inputBufferSize    = 10000
	inputBufferWarning = 0.3

	messageBufferSize    = 100
	messageBufferWarning = 0.3

	errorBufferSize    = 100
	errorBufferWarning = 0.3

	stdinBufferStartCapacity = 32 * 1024
	stdinBufferMaxCapacity   = 2048 * 1024
)

// MessageHandler is called for every parsed message on the message channel
type MessageHandler func(msg Message) error

// ErrorHandler is called for every error on the error channel
type ErrorHandler func(err error)

// Reader holds runtime config and information about the reader instance
type Reader struct {
	messageChan     chan Message
	errorChan       chan error
	inputBufferChan chan string
	MessageHandler  MessageHandler
	ErrorHandler    ErrorHandler
}

// Message a contains the content and metadata of a channel message
type Message struct {
	Type        string
	Timestamp   time.Time
	Test        string
	Sender      net.IP
	Hostname    string
	HostAddress net.IP
	Color       string
	OldColor    string
	LastChange  time.Time
	Page        string
	OSName      string
	ClassName   string
	Body        []string
}

func parseTimestamp(timestamp string) (time.Time, error) {
	fields := strings.Split(timestamp, ".")
	if len(fields) != 2 {
		return time.Time{}, errors.New("Malformed timestamp")
	}
	seconds, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	useconds, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, useconds), nil
}

// ParseMessage parses a message into the Message type
func (r *Reader) ParseMessage(msg []string) {

	var (
		err     error
		message Message
	)

	for _, line := range msg {

		// If the line starts with "@@", it's the header. Parse it and fill in the metadata fields.
		// Otherwise append the line to the message body
		if strings.HasPrefix(line, "@@") {

			fields := strings.Split(line, "|")
			if strings.HasPrefix(line, "@@page") {
				if len(fields) != 16 {
					r.errorChan <- fmt.Errorf("Malformed message header: %+v", line)
					return
				}
				message.Type = "page"
				message.LastChange, err = parseTimestamp(fmt.Sprintf("%s.0", fields[9]))
				if err != nil {
					r.errorChan <- err
					return
				}
				message.Color = fields[7]
				message.OldColor = fields[8]
				message.Page = fields[10]
				message.OSName = fields[12]
				message.ClassName = fields[13]
			} else if strings.HasPrefix(line, "@@ack") {
				if len(fields) != 7 {
					r.errorChan <- errors.New("Malformed message header")
					return
				}
				message.Type = "ack"
			} else {
				r.errorChan <- errors.New("Unknown message type. Raw header: " + line)
				return
			}

			message.Timestamp, err = parseTimestamp(fields[1])
			if err != nil {
				r.errorChan <- err
				return
			}

			message.Sender = net.ParseIP(fields[2])
			message.Hostname = fields[3]
			message.Test = fields[4]
			message.HostAddress = net.ParseIP(fields[5])
			// field 6 is discarded in favor of the timestamp in 1

		} else {
			message.Body = append(message.Body, line)
		}
	}

	r.messageChan <- message
}

func (r *Reader) messageWorker() {
	for msg := range r.messageChan {
		err := r.MessageHandler(msg)
		if err != nil {
			r.errorChan <- err
		}
	}
}

func (r *Reader) errorWorker() {
	for msg := range r.errorChan {
		r.ErrorHandler(msg)
	}
}

func (r *Reader) logDebugf(msg string, a ...interface{}) {
	if os.Getenv("GOXYMON_DEBUG") == "true" {
		log.Printf("XymonChannelReader: "+msg, a...)
	}
}

func (r *Reader) logInfof(msg string, a ...interface{}) {
	log.Printf("XymonChannelReader: "+msg, a...)
}

// Run starts the Stdin reader
func (r *Reader) Run() {
	r.logDebugf("STDIN reader starting")
	stdin := bufio.NewScanner(os.Stdin)
	// Adjust the scan buffer size for large check bodies
	buf := make([]byte, stdinBufferStartCapacity)
	stdin.Buffer(buf, stdinBufferMaxCapacity)

	for stdin.Scan() {
		r.inputBufferChan <- stdin.Text()
	}
}

func (r *Reader) bufferDispatcher() {
	//headerRegex := regexp.MustCompile(`^@@\w`)
	var currentMessage []string
	startTime := time.Time{}

	for line := range r.inputBufferChan {
		// If this is a new Message, clear currentMessage
		if (strings.HasPrefix(line, "@@")) && (line != "@@") {
			currentMessage = nil
			startTime = time.Now()
		}
		// If this is the end of a message, dispatch a parser goroutine and carry on
		if line == "@@" {
			diff := time.Now().Sub(startTime)
			r.logDebugf("Read %s message in %s",
				humanize.Bytes(uint64(len(currentMessage))), diff)
			go r.ParseMessage(currentMessage)
		}

		currentMessage = append(currentMessage, line)
	}
}

func (r *Reader) queueMonitor() {
	for {
		time.Sleep(3000 * time.Millisecond)
		length := len(r.inputBufferChan)
		r.logDebugf("Input buffer length: %d", length)
		if float64(length) >= float64(cap(r.inputBufferChan))*inputBufferWarning {
			r.logInfof("Input buffer length over threshold!")
		}
		length = len(r.errorChan)
		r.logDebugf("Error queue length: %d", length)
		if float64(length) >= float64(cap(r.errorChan))*errorBufferWarning {
			r.logInfof("Error queue length over threshold!")
		}
		length = len(r.messageChan)
		r.logDebugf("Message queue length: %d", length)
		if float64(length) >= float64(cap(r.messageChan))*messageBufferWarning {
			r.logInfof("Message queue length over threshold!")
		}
	}
}

// NewReader sets up a channel reader which reads mesages from STDIN, parses them and dispatches them to the handler functions.
// Every parsed message will be fed to the MessageHandler.
// Every error will be fed to the ErrorHandler.
func NewReader(messageHandler MessageHandler, errorHandler ErrorHandler) *Reader {
	reader := Reader{
		inputBufferChan: make(chan string, inputBufferSize),
		messageChan:     make(chan Message, messageBufferSize),
		errorChan:       make(chan error, errorBufferSize),
		MessageHandler:  messageHandler,
		ErrorHandler:    errorHandler,
	}

	go reader.queueMonitor()
	reader.logDebugf("Queue monitor started")
	go reader.bufferDispatcher()
	reader.logDebugf("Input buffer dispatcher started")
	go reader.messageWorker()
	reader.logDebugf("Message worker started")
	go reader.errorWorker()
	reader.logDebugf("Error message worker started")

	return &reader
}
