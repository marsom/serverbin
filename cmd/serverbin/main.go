package main

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marsom/serverbin/cmd/serverbin/cmd"
)

//nolint:gochecknoglobals //this vars are set on build by goreleaser
var (
	version = "0.0.0"
	commit  = "12345678"
	date    = "2020-09-23T07:03:55+02:00"
)

type logWriter struct {
}

func (l logWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(time.Now().Format("2006-01-02T15:04:05.000Z07:00") + " " + string(p))
}

// make log statements with meaningful timestamps
func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

var cli struct {
	HttpCmd    cmd.HttpCmd `kong:"cmd,name='http',help='Start a HTTP test server'"`
	TcpCmd     cmd.TcpCmd  `kong:"cmd,name='tcp',help='Start a TCP test server'"`
	VersionCmd versionCmd  `kong:"cmd,name='version',help='Print version information'"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.TypeMapper(reflect.TypeOf(&net.IPNet{}), ipnetMapper()),
	)
	ctx.FatalIfErrorf(ctx.Run())
}

func ipnetMapper() kong.MapperFunc {
	return func(ctx *kong.DecodeContext, target reflect.Value) error {
		var value string
		if err := ctx.Scan.PopValueInto("ipnet", &value); err != nil {
			return err
		}

		_, ipnet, err := net.ParseCIDR(value)
		if err != nil {
			return fmt.Errorf("expected ipnet but got %q: %w", value, err)
		}

		target.Set(reflect.ValueOf(ipnet))

		return nil
	}
}

type versionCmd struct {
}

func (cmd *versionCmd) Run() error {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("BuildDate: %s\n", date)
	fmt.Printf("GoVersion: %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}
