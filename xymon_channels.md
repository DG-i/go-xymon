# Xymon Channels
This documentation is based on the [source code of Xymon 4.3.28][1]

Each channel receives specific message types. Additionally, each channel receives the following message types:

* `drophost|timestamp|sender|hostname`
* `dropstate|timestamp|sender|hostname`
* `droptest|timestamp|sender|hostname|testname`
* `renamehost|timestamp|sender|hostname|newhostname`
* `renametest|timestamp|sender|hostname|oldtestname|newtestname`
* `reload|timestamp|sender`
* `shutdown|timestamp|sender`
* `logrotate|timestamp|sender`
* `idle|timestamp|sender`

## Available channels

### Page channel
This channel is fed information about tests where the color changes between an alert color and a non-alert color. It also receives information about "ack" messages. Possible message types:
* ack
* notify
* page

#### Header examples
```
@@page#472566/foo.example.com|1497700544.139090|1.1.1.1|foo.example.com|apt|1.1.1.1|1497702344|red|red|1487820721|examples|476945|linux|linux||
@@page#667449/bar.example.com|1505028445.483655|2.2.2.2|bar.example.com|pkg|2.2.2.2|1505039245|red|red|1504142860|examples|423156|freebsd|freebsd||
@@page#472554/baz.example.com|1497700539.738164|3.3.3.3|baz.example.com|updates|3.3.3.3|1497702339|red|red|1496146101|examples|437203||||
@@ack#665787/foo.domain.com|1505027914.438057|10.10.10.10|foo.domain.com|apt|4.4.4.4|1505028214
```

#### Header fields
##### "page" messages

| field | value                      |
|-------|----------------------------|
| 0     | channel marker with sender |
| 1     | microsecond timestamp      |
| 2     | sender IP                  |
| 3     | sender hostname            |
| 4     | testname                   |
| 5     | host IP                    |
| 6     | test expiration            |
| 7     | check color                |
| 8     | old check color            |
| 9     | last change timestamp      |
| 10    | page path                  |
| 11    | cookie                     |
| 12    | OS name                    |
| 13    | class name                 |
| 14    | grouplist                  |

##### "ack" messages

| field | value                      |
|-------|----------------------------|
| 0     | channel marker with sender |
| 1     | microsecond timestamp      |
| 2     | sender IP                  |
| 3     | sender hostname            |
| 4     | testname                   |
| 5     | host IP                    |
| 6     | ack expiration             |


##### "notify" messages

| field | value                      |
|-------|----------------------------|
| 0     | channel marker with sender |
| 1     | microsecond timestamp      |
| 2     | sender IP                  |
| 3     | sender hostname            |
| 4     | testname                   |
| 5     | page path                  |

### Status channel
This channel is fed the contents of all incoming "status" and "summary" messages.
#### Header example
#### Header fields

### Stachg channel
This channel is fed information about tests that change status, i.e. the color of the status-log changes.
#### Header example
#### Header fields

### Data channel
This channel is fed information about all "data" messages.
#### Header example
#### Header fields

### Notes channel
This channel is fed information about all "notes" messages.
#### Header example
#### Header fields

### Enadis channel
This channel is fed information about hosts or tests that are being disabled or enabled.
#### Header example
#### Header fields

### Client channel
This channel is fed the contents of the client messages sent by Xymon clients installed on the monitored servers.
#### Header example
#### Header fields

### Clichg channel
This channel is fed the contents of a host client messages, whenever a status for that host goes red, yellow or purple.
#### Header example
#### Header fields

[1]: https://sourceforge.net/p/xymon/code/HEAD/tree/trunk/xymond/xymond.c
