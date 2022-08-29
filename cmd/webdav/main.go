package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/webdav"
)

func init() {
	pflag.StringP("resource-root", "r", ".", "webdav resource root dir")
	pflag.Uint16P("port", "p", 8080, "webdav server listen port")
	pflag.Parse()

	_ = viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
}

func main() {
	dav := &webdav.Handler{
		Prefix:     "",
		FileSystem: webdav.Dir(viper.GetString("resource-root")),
		LockSystem: webdav.NewMemLS(),
		Logger: func(request *http.Request, err error) {
			if err != nil {
				logrus.Error(err)
				buf, _ := httputil.DumpRequest(request, false)
				logrus.Infof("%s", buf)
			}
		},
	}
	listenAddr := fmt.Sprintf(":%d", viper.GetUint("port"))
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	defer l.Close()
	logrus.Infof("listem %s", listenAddr)

	s := http.Server{Handler: dav}
	err = s.Serve(l)
	if err != nil {
		logrus.Fatal(err)
	}
}
