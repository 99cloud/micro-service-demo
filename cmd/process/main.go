package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"demo/pkg/utils"
)

func init() {
	pflag.Bool("grayscale", false, "grayscale output image")
	pflag.String("webdav-endpoint", "", "webdav server endpoint")
	pflag.String("watermark", "", "watermark text")
	pflag.Uint16("port", 8080, "listen port")
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logrus.Fatal(err)
	}
	viper.AutomaticEnv()
	viper.BindEnv()
}

func main() {
	listenAddr := fmt.Sprintf(":%d", viper.GetUint("port"))
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	defer l.Close()
	logrus.Infof("listen %s", listenAddr)
	u, err := utils.ParseUrl(viper.GetString("webdav-endpoint"))
	if err != nil {
		logrus.Fatal("invalid webdav url")
	}
	h := process{storage: gowebdav.NewClient(u.String(), "", "")}
	s := http.Server{Handler: h}
	err = s.Serve(l)
	if err != nil {
		logrus.Fatal(err)
	}
}

type process struct {
	storage *gowebdav.Client
}

func (p process) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/_statusCode_" {
		utils.HTTPStatusCode(writer, request)
		return
	}
	switch request.Method {
	case http.MethodGet:
		p.LoadFile(writer, request)
	case http.MethodPut:
		p.SaveFile(writer, request)
	default:
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (p process) SaveFile(writer http.ResponseWriter, request *http.Request) {
	img, _, err := image.Decode(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	img = imaging.Resize(img, 0, 800, imaging.NearestNeighbor)
	buf := &bytes.Buffer{}
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	err = p.storage.WriteStream(request.URL.Path, buf, 0)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
}

func (p process) LoadFile(writer http.ResponseWriter, request *http.Request) {
	imgData, err := p.storage.ReadStream(request.URL.Path)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	defer imgData.Close()
	watermark := viper.GetString("watermark")
	gray := viper.GetBool("grayscale")
	if !gray && watermark == "" {
		_, _ = io.Copy(writer, imgData)
		return
	}
	img, _, err := image.Decode(imgData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if gray {
		img = imaging.Grayscale(img)
	}
	if watermark != "" {
		if watermark == "$POD" {
			watermark = utils.GetPodName()
		}
		dimg, ok := img.(draw.Image)
		if !ok {
			dimg = imaging.Clone(img)
		}
		err = DrawText(dimg, watermark, 20, image.Point{X: 50, Y: 750})
		if err != nil {
			logrus.Errorf("DrawText:%s", err)
		}
	}
	_ = jpeg.Encode(writer, img, &jpeg.Options{Quality: 75})
}

type dav struct {
	endpoint string
}
