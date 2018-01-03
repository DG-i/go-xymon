package channels

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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

func parsePageHeader(fields []string) (msg Message, err error) {

	if len(fields) != 16 {
		return msg, fmt.Errorf("Malformed message header: %+v", fields)
	}
	msg.Type = TypePage
	msg.LastChange, err = parseTimestamp(fmt.Sprintf("%s.0", fields[9]))
	if err != nil {
		return msg, err
	}
	msg.Timestamp, err = parseTimestamp(fields[1])
	if err != nil {
		return msg, err
	}

	msg.Color = fields[7]
	msg.OldColor = fields[8]
	msg.Page = fields[10]
	msg.OSName = fields[12]
	msg.ClassName = fields[13]

	msg.Sender = net.ParseIP(fields[2])
	msg.Hostname = fields[3]
	msg.Test = fields[4]
	msg.HostAddress = net.ParseIP(fields[5])

	return msg, nil
}

func parseAckHeader(fields []string) (msg Message, err error) {

	if len(fields) != 7 {
		return msg, fmt.Errorf("Malformed message header: %+v", fields)
	}
	msg.Type = TypeAck

	msg.Timestamp, err = parseTimestamp(fields[1])
	if err != nil {
		return msg, err
	}

	msg.Sender = net.ParseIP(fields[2])
	msg.Hostname = fields[3]
	msg.Test = fields[4]
	msg.HostAddress = net.ParseIP(fields[5])

	return msg, nil
}
