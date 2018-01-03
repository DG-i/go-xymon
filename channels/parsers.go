package channels

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Message a contains the content and metadata of a channel message
type Message struct {
	Type               string
	Timestamp          time.Time
	Test               string
	NewTest            string
	Sender             net.IP
	Hostname           string
	NewHostname        string
	HostAddress        net.IP
	Color              string
	OldColor           string
	LastChange         time.Time
	Page               string
	OSName             string
	ClassName          string
	ExpireTime         time.Time
	AckExpire          time.Time
	AckMessage         string
	DisableExpire      time.Time
	DisableMessage     string
	Body               []string
	ClientMsgTimestamp time.Time
	Flapping           bool
	DowntimeActive     bool
	Modifiers          string
}

func checkFieldCount(fields []string, desired int) error {
	if len(fields) != desired {
		return fmt.Errorf("Malformed message header: %+v", fields)
	}
	return nil
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

func (msg *Message) parseCommonFields(fields []string) (err error) {
	msg.Timestamp, err = parseTimestamp(fields[1])
	if err != nil {
		return err
	}
	msg.Sender = net.ParseIP(fields[2])
	if len(fields) >= 4 {
		msg.Hostname = fields[3]
	}
	return nil
}

func (msg *Message) parseAckHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 7); err != nil {
		return err
	}
	msg.Test = fields[4]
	msg.HostAddress = net.ParseIP(fields[5])
	return nil
}

func (msg *Message) parseEnaDisHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 7); err != nil {
		return err
	}
	msg.Test = fields[4]
	msg.DisableExpire, err = parseTimestamp(fields[5])
	if err != nil {
		return err
	}
	msg.DisableMessage = fields[6]
	return nil
}

func (msg *Message) parseDataHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 8); err != nil {
		return err
	}

	msg.Timestamp, err = parseTimestamp(fields[1])
	if err != nil {
		return err
	}
	msg.Sender = net.ParseIP(fields[2])
	msg.Hostname = fields[4]
	msg.Test = fields[5]
	msg.ClassName = fields[6]
	msg.Page = fields[7]
	return nil
}

func (msg *Message) parseNotifyHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 6); err != nil {
		return err
	}
	msg.Test = fields[4]
	msg.Page = fields[5]
	return nil
}

func (msg *Message) parsePageHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 16); err != nil {
		return err
	}

	msg.LastChange, err = parseTimestamp(fmt.Sprintf("%s.0", fields[9]))
	if err != nil {
		return err
	}

	msg.Color = fields[7]
	msg.OldColor = fields[8]
	msg.Page = fields[10]
	msg.OSName = fields[12]
	msg.ClassName = fields[13]
	msg.Test = fields[4]
	msg.HostAddress = net.ParseIP(fields[5])

	return nil
}

func (msg *Message) parseStaChgHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 15); err != nil {
		return err
	}

	msg.Timestamp, err = parseTimestamp(fields[1])
	msg.ExpireTime, err = parseTimestamp(fields[6])
	msg.LastChange, err = parseTimestamp(fields[9])
	msg.DisableExpire, err = parseTimestamp(fields[10])
	msg.ClientMsgTimestamp, err = parseTimestamp(fields[13])
	if err != nil {
		return err
	}
	msg.Sender = net.ParseIP(fields[2])
	msg.Hostname = fields[4]
	msg.Test = fields[5]
	msg.Color = fields[7]
	msg.OldColor = fields[8]
	msg.DisableMessage = fields[11]
	if fields[12] == "0" {
		msg.DowntimeActive = false
	} else {
		msg.DowntimeActive = true
	}
	msg.Modifiers = fields[14]
	return nil
}

func (msg *Message) parseStatusHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 20); err != nil {
		return err
	}

	msg.Timestamp, err = parseTimestamp(fields[1])
	msg.ExpireTime, err = parseTimestamp(fields[6])
	msg.LastChange, err = parseTimestamp(fmt.Sprintf("%s.0", fields[10]))
	msg.AckExpire, err = parseTimestamp(fmt.Sprintf("%s.0", fields[11]))
	msg.DisableExpire, err = parseTimestamp(fmt.Sprintf("%s.0", fields[13]))
	msg.ClientMsgTimestamp, err = parseTimestamp(fmt.Sprintf("%s.0", fields[15]))
	if err != nil {
		return err
	}

	msg.Sender = net.ParseIP(fields[2])
	msg.Hostname = fields[4]
	msg.Test = fields[5]
	msg.Color = fields[7]
	msg.OldColor = fields[9]
	msg.AckMessage = fields[12]
	msg.DisableMessage = fields[14]
	msg.ClassName = fields[16]
	msg.Page = fields[17]
	if fields[18] == "0" {
		msg.Flapping = false
	} else {
		msg.Flapping = true
	}
	msg.Modifiers = fields[19]

	return nil
}

func (msg *Message) parseDropTestHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 5); err != nil {
		return err
	}
	msg.Test = fields[4]
	return nil
}

func (msg *Message) parseRenameHostHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 5); err != nil {
		return err
	}
	msg.NewHostname = fields[4]
	return nil
}

func (msg *Message) parseRenameTestHeader(fields []string) (err error) {
	if err := checkFieldCount(fields, 6); err != nil {
		return err
	}
	msg.Test = fields[4]
	msg.NewTest = fields[5]
	return nil
}
