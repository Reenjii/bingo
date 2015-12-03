# bingo
A pastebin-like written in Go where the server has zero knowledge of pasted data. Data is encrypted/decrypted in the browser using AES.

## Build

bingo is written in go so it compiles easily through `go install github.com/reenjii/bingo/cmd/bingo`.

Assets (templates, scripts, styles, conf) are managed by [Grunt](http://www.gruntjs.com/).
[Install Grunt](http://www.gruntjs.com/installing-grunt) and run `grunt` to build assets in the `dist` directory.

## Run

The server uses a json configuration file (defaults to `/etc/bingo.json`). The configuration path can be set in the command line: `$GOPATH/bin/bingo -conf /path/to/my/bingo.json`.
An example configuration file is deployed in `dist/conf` folder by grunt.

The configuration file contains the path of the `views` and the `assets` (`static` folder). You must set these paths to make data in `dist/views` and in `dist/static` available to the server.

## Example

```
$> $GOPATH/bin/bingo                                                                                                                                                                                                              [1]
INFO  2015/12/03 11:35:06.064178 Bingo initialization
INFO  2015/12/03 11:35:06.064682 Templates initialization
INFO  2015/12/03 11:35:06.064820 Views are in  /tmp/bingo/dist/views/*.html
INFO  2015/12/03 11:35:06.064880 Loading template /tmp/bingo/dist/views/paste.html
INFO  2015/12/03 11:35:06.065444 Create folder /tmp/bingo/data
INFO  2015/12/03 11:35:06.065530 Build paste index...
INFO  2015/12/03 11:35:06.065634 Paste index built with 0 entries
INFO  2015/12/03 11:35:06.065697 Start clean daemon with a 3600 seconds threshold
INFO  2015/12/03 11:35:06.065823 Listening on :1337
```
