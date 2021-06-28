# Pterodactyl Packet Watch

## Description
A project based off of my Pterodactyl Game Server Watch [tool](https://github.com/gamemann/Pterodactyl-Game-Server-Watch). This specific project basically sends specified packet types and if the average response time (in milliseconds) goes beyond a specified threshold, misc options (web hooks) will be performed.

## Command Line Flags
There is only one command line argument/flag and it is `-cfg=<path>`. This argument/flag changes the path to the Pteropckt config file. The default value is `/etc/pteropckt/pteropckt.conf`.

Examples include:

```
./pteropckt -cfg=/home/cdeacon/myconf.conf
./pteropckt -cfg=~/myconf.conf
./pteropckt -cfg=myconf.conf
```

## Config File
The config file's default path is `/etc/pteropckt/pteropckt.conf` (this can be changed with a command line argument/flag as seen above). This should be a JSON array including the API URL, token, and an array of servers to check against. The main options are the following:

* `apiurl` => The Pterodactyl API URL (do not include the `/` at the end).
* `token` => The bearer token (from the client) to use when sending requests to the Pterodactyl API.
* `apptoken` => The bearer token (from the application) to use when sending requests to the Pterodactyl API (this is only needed when `addservers` is set to `true`).
* `debug` => The debug level (1-4).
* `reloadtime` => If above 0, will reload the configuration file and retrieve servers from the API every *x* seconds.
* `addservers` => Whether or not to automatically add servers to the config from the Pterodactyl API.
* `defenable` => The default enable boolean of a server added via the Pterodactyl API.
* `defthreshold` => The default threshold of a server added via the Pterodactyl API.
* `defcount` => The default count (max latencies stored) of a server added via the Pterodactyl API.
* `definterval` => The default interal between scanning servers of a server added via the Pterodactyl API.
* `deftimeout` => The default packet timeout of a server added via the Pterodactyl API.
* `defmaxdetects` => The default max detects of a server added via the Pterodactyl API.
* `defcooldown` => The default cooldown between detects of a server added via the Pterodactyl API.
* `defmentions` => The default mentions JSON for servers added via the Pterodactyl API.
* `servers` => An array of servers to watch (read below).
* `misc` => An array of misc options (read below).

## Egg Variable Overrides
If you have the `addservers` setting set to true (servers are automatically retrieved via the Pterodactyl API), you may use the following egg variables as overrides to the specific server's config.

* `PTEROPCKT_DISABLE` => If set to above 0, will disable the specific server from the tool.
* `PTEROPCKT_IP` => If not empty, will override the server IP to scan with this value for the specific server.
* `PTEROPCKT_PORT` => If not empty, will override the server port to scan with this value for the specific server.
* `PTEROPCKT_THRESHOLD` => If not empty, will override the threshold with this value for the specific server.
* `PTEROPCKT_COUNT` => If not empty, will override the count with this value for the specific server.
* `PTEROPCKT_INTERVAL` => If not empty, will override the interval with this value for the specific server.
* `PTEROPCKT_TIMEOUT` => If not empty, will override the timeout with this value for the specific server.
* `PTEROPCKT_MAXDETECTS` => If not empty, will override max detects with this value for the specific server.
* `PTEROPCKT_COOLDOWN` => If not empty, will override cooldown with this value for the specific server.
* `PTEROPCKT_MENTIONS` => If not empty, will override the mentions JSON string with this value for the specific server.

## Server Options/Array
This array is used to manually add servers to watch. The `servers` array should contain the following items:

* `name` => The server's name.
* `enable` => If true, this server will be scanned.
* `ip` => The IP to send A2S_INFO requests to.
* `port` => The port to send A2S_INFO requests to.
* `uid` => The server's Pterodactyl UID.
* `threshold` => The default threshold of the server.
* `count` => The default count (max latencies stored) of the server.
* `interval` => The default interal between scanning servers of the server.
* `timeout` => The default packet timeout of the server.
* `maxdetects` => The default max detects of the server.
* `cooldown` => The default cooldown between detects of the server.
* `packets` => The packets array.
* `mentions` => A JSON string that parses all custom role and user mentions inside of web hooks for this server.

## Packets Array
The server's `packets` array defines which packets the servers should send and the main point of the program.

* `name` => The server's name.
* `data` => The request payload in hexadecimal format with no spaces.
* `threshold` => The threshold of the server.
* `count` => The count (max latencies stored) of the server.
* `interval` => The interal between scanning servers of the server.
* `timeout` => The packet timeout of the server.
* `maxdetects` => The max detects of the server.
* `cooldown` => The cooldown between detects of the server.

The following is an example.

```JSON
"packets": [
        {
                "name": "name",
                "data": "FFFFFFFF54536F7572636520456E67696E6520517565727900"
        }
]
```

## Server Mentions Array
The server `mentions` JSON string's parsed JSON output includes a `data` list with each item including a `role` (boolean indicating whether we're mentioning a role) and `id` (the ID of the role or user in string format).

Here are some examples:

```JSON
{
        "data": [
                {
                        "role": true,
                        "id": "1293959919293959192"
                },
                {
                        "role": false,
                        "id": "1959192351293954123"
                }
        ]
}
```

This is what it looks like inside of the mentions string.

```JSON
{
        "servers": [
                {
                        "mentions": "{\"data\":[{\"role\": true,\"id\": \"1293959919293959192\"},{\"role\": false,\"id\": \"1959192351293954123\"}]}"
                }
        ]
}
```

The above will replace the `{MENTIONS}` text inside of the web hook's contents with `<@&1293959919293959192>, <@1959192351293954123>`.

## Misc Options/Array
This tool supports misc options which are configured under the `misc` array inside of the config file. The only event supported for this at the moment is when a server is restarted from the tool. However, other events may be added in the future. An example may be found below.

```JSON
{
        "misc": [
                {
                        "type": "misctype",
                        "data": {
                                "option1": "val1",
                                "option2": "val2"
                        }
                }
        ]
}
```

### Web Hooks
As of right now, the only misc option `type` is `webhook` which indicates a web hook. The `app` data item represents what type of application the web hook is for (the default value is `discord`).

Please look at the following data items:

* `app` => The web hook's application (either `discord` or `slack`).
* `url` => The web hook's URL (**REQUIRED**).
* `contents` => The contents of the web hook.
* `username` => The username the web hook sends as (**only** Discord).
* `avatarurl` => The avatar URL used with the web hook (**only** Discord).
* `mentions` => An array including a `roles` item as a boolean allowing custom role mentions and `users` item as a boolean allowing custom user mentions.

**Note** - Please copy the full web hook URL including `https://...`.

#### Variable Replacements For Contents
The following strings are replaced inside of the `contents` string before the web hook submission.

* `{IP}` => The server's IP.
* `{PORT}` => The server's port.
* `{NAME}` => The server's name.
* `{UID}` => The server's UID from the config file/Pterodactyl API.
* `{AVG}` => The server's average latency.
* `{MAX}` => The server's max latency.
* `{MIN}` => The server's min latency.
* `{THRESHOLD}` => The packet's configured threshold.
* `{COUNT}` => The packet's configured count.
* `{INTERVAL}` => The packet's configured interval.
* `{TIMEOUT}` => The packet's configured timeout.
* `{MAXDETECTS}` => The packet's configured max detects.
* `{COOLDOWN}` => The packet's configured cooldown.
* `{MENTIONS}` => If there are mentions, it will print them in `<id>, ...` format in this replacement.

#### Defaults
Here are the Discord web hook's default values.

* `contents` => ...
* `username` => Pteropckt
* `avatarurl` => *empty* (default)

## Configuration Example
You may find config examples in the [tests/](https://github.com/gamemann/Pterodactyl-Packet-Watch/tree/master/tests) directory.

## Building
You may use `git` and `go build` to build this project and produce a binary. I'd suggest cloning this to `$GOPATH` so there aren't problems with linking modules. For example:

```
cd <Path To One $GOPATH>
git clone https://github.com/gamemann/Pterodactyl-Packet-Watch.git
cd Pterodactyl-PacketWatch
go build -o pteropckt
```

## Credits
* [Christian Deacon](https://github.com/gamemann) - Creator.