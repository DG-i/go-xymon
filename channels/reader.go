package channels // import "github.com/dg-i/go-xymon/channels"

import (
	"bufio"
	"errors"
	"log"
	"os"
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

// Possible message types
const (
	// all channels
	TypeDropHost   = "drophost"
	TypeDropState  = "dropstate"
	TypeDropTest   = "droptest"
	TypeRenameHost = "renamehost"
	TypeRenameTest = "renametest"
	TypeReload     = "reload"
	TypeShutdown   = "shutdown"
	TypeLogrotate  = "logrotate"
	TypeIdle       = "idle"

	// enadis channel
	TypeEnaDis = "enadis"

	// data channel
	TypeData = "data"

	// notes channel
	TypeNotes = "notes"

	// page channel
	TypeAck    = "ack"
	TypeNotify = "notify"
	TypePage   = "page"

	// stachg channel
	TypeStaChg = "stachg"

	// status channel
	TypeStatus = "status"
)

// Handlers defines the interface for message handlers
type Handlers interface {
	// MessageHandler is called for every parsed message on the message channel
	MessageHandler(msg Message) error
	// ErrorHandler is called for every error on the error channel
	ErrorHandler(err error)
}

// Reader holds runtime config and information about the reader instance
type Reader struct {
	messageChan     chan Message
	errorChan       chan error
	inputBufferChan chan string
	handlers        Handlers
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
			// Header should look like @@page#472566/foo.example.com|...
			msgType := strings.Trim(strings.Split(fields[0], "#")[0], "@")

			switch msgType {
			case TypeAck:
				message.Type = TypeAck
				err = message.parseCommonFields(fields)
				err = message.parseAckHeader(fields)

			case TypeEnaDis:
				message.Type = TypeEnaDis
				err = message.parseCommonFields(fields)
				err = message.parseEnaDisHeader(fields)

			case TypeData:
				message.Type = TypeData
				err = message.parseDataHeader(fields)

			case TypeNotify:
				message.Type = TypeNotify
				err = message.parseCommonFields(fields)
				err = message.parseNotifyHeader(fields)

			case TypePage:
				message.Type = TypePage
				err = message.parseCommonFields(fields)
				err = message.parsePageHeader(fields)

			case TypeStaChg:
				message.Type = TypeStaChg
				err = message.parseStaChgHeader(fields)

			case TypeStatus:
				message.Type = TypeStatus
				err = message.parseStatusHeader(fields)

			case TypeNotes, TypeReload, TypeShutdown, TypeLogrotate, TypeIdle, TypeDropHost, TypeDropState:
				message.Type = msgType
				err = message.parseCommonFields(fields)

			case TypeDropTest:
				message.Type = TypeDropTest
				err = message.parseCommonFields(fields)
				err = message.parseDropTestHeader(fields)

			case TypeRenameHost:
				message.Type = TypeRenameHost
				err = message.parseCommonFields(fields)
				err = message.parseRenameHostHeader(fields)

			case TypeRenameTest:
				message.Type = TypeRenameTest
				err = message.parseCommonFields(fields)
				err = message.parseRenameTestHeader(fields)

			default:
				err = errors.New("Unknown message type. Raw header: " + line)
			}

			if err != nil {
				r.errorChan <- err
				return
			}

		} else {
			message.Body = append(message.Body, line)
		}
	}

	r.messageChan <- message
}

func (r *Reader) messageWorker() {
	for msg := range r.messageChan {
		err := r.handlers.MessageHandler(msg)
		if err != nil {
			r.errorChan <- err
		}
	}
}

func (r *Reader) errorWorker() {
	for msg := range r.errorChan {
		r.handlers.ErrorHandler(msg)
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

		// We're not breaking the loop above, so the last "@@" line will be appended to
		// currentMessage. This doesn't matter as currentMessage will be reset on the next
		// iteration.
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
func NewReader(handlers Handlers) *Reader {
	reader := Reader{
		inputBufferChan: make(chan string, inputBufferSize),
		messageChan:     make(chan Message, messageBufferSize),
		errorChan:       make(chan error, errorBufferSize),
		handlers:        handlers,
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
