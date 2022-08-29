package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"

	"demo/pkg/utils"
)

func init() {
	pflag.String("webdav-endpoint", "", "webdav server endpoint")
	pflag.String("process-endpoint", "", "process server endpoint")
	pflag.Uint16("port", 8080, "listen port")
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	if err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	listenAddr := fmt.Sprintf(":%d", viper.GetUint("port"))
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	defer l.Close()
	logrus.Infof("listen %s", listenAddr)

	s := http.Server{Handler: imageManager{
		webdavEndpoint:  viper.GetString("webdav-endpoint"),
		processEndpoint: viper.GetString("process-endpoint"),
	}}
	err = s.Serve(l)
	if err != nil {
		logrus.Fatal(err)
	}
}

type imageManager struct {
	webdavEndpoint  string
	processEndpoint string
}

func (m imageManager) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	utils.InjectPodName(writer.Header())
	switch request.URL.Path {
	case "/_statusCode_":
		utils.HTTPStatusCode(writer, request)
		return
	case "/process/_statusCode_":
		u, err := utils.ParseUrl(m.processEndpoint)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		u.Path = path.Join("_statusCode_")
		u.RawQuery = request.URL.RawQuery
		resp, err := http.Get(u.String())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		writer.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(writer, resp.Body)
		return
	}

	if request.URL.Path == "/images" {
		switch request.Method {
		case http.MethodGet:
			m.listImages(writer, request)
		case http.MethodPut, http.MethodPost:
			m.uploadImage(writer, request)
		default:
			http.Error(writer,
				http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(request.URL.Path, "/images") {
		switch request.Method {
		case http.MethodGet:
			m.getImage(writer, request)
		case http.MethodDelete:
			m.deleteImage(writer, request)
		default:
			http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	} else {
		http.FileServer(http.Dir("html")).ServeHTTP(writer, request)
	}
}

func (m imageManager) getImage(writer http.ResponseWriter, request *http.Request) {
	imgPath := strings.TrimPrefix(request.URL.Path, "/images")
	imgData, code, err := m.ReadImage(imgPath, request.Header)
	if err != nil {
		http.Error(writer, err.Error(), code)
		return
	}
	_, _ = io.Copy(writer, imgData)
	_ = imgData.Close()
}

func (m imageManager) deleteImage(writer http.ResponseWriter, request *http.Request) {
	imgPath := strings.TrimPrefix(request.URL.Path, "/images")
	dav, err := m.webdav()
	if err != nil {
		http.Error(writer, err.Error(),
			http.StatusInternalServerError)
		return
	}
	err = dav.Remove(imgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(writer, request)
		} else {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("success"))
}

func (m imageManager) uploadImage(writer http.ResponseWriter, request *http.Request) {
	imgPath := fmt.Sprintf("/%s.jpeg", strconv.FormatInt(time.Now().UnixNano(), 36))
	err := m.SaveImage(imgPath, request.Header, request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

func (m imageManager) webdav() (*gowebdav.Client, error) {
	u, err := utils.ParseUrl(m.webdavEndpoint)
	if err != nil {
		return nil, err
	}
	dc := gowebdav.NewClient(u.String(), "", "")
	return dc, err
}

func (m imageManager) listImages(writer http.ResponseWriter, _ *http.Request) {
	dav, err := m.webdav()
	if err != nil {
		http.Error(writer, err.Error(),
			http.StatusInternalServerError)
		return
	}
	fs, err := dav.ReadDir("/")
	if err != nil {
		http.Error(writer, err.Error(),
			http.StatusInternalServerError)
		return
	}
	imgs := make([]string, 0, len(fs))
	for _, fi := range fs {
		name := fi.Name()
		if strings.HasSuffix(name, ".jpeg") {
			imgs = append(imgs, "/images/"+name)
		}
	}
	_ = json.NewEncoder(writer).Encode(imgs)
}

func (m imageManager) SaveImage(key string, header http.Header, r io.Reader) error {
	u, err := utils.ParseUrl(m.processEndpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, key)
	req, err := http.NewRequest(http.MethodPut, u.String(), r)
	if err != nil {
		return err
	}
	utils.CopyHeader(req.Header, header)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		buf := make([]byte, 8196)
		n, _ := resp.Body.Read(buf)
		buf = buf[:n]
		return fmt.Errorf("http code %d body: %s", resp.StatusCode, buf)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (m imageManager) ReadImage(key string, header http.Header) (io.ReadCloser, int, error) {
	u, err := utils.ParseUrl(m.processEndpoint)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	u.Path = path.Join(u.Path, key)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	utils.CopyHeader(req.Header, header)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if resp.StatusCode != http.StatusOK {
		buf := make([]byte, 8196)
		n, _ := resp.Body.Read(buf)
		buf = buf[:n]
		_ = resp.Body.Close()
		return nil, resp.StatusCode, fmt.Errorf("http code %d body: %s", resp.StatusCode, buf)
	}
	return resp.Body, http.StatusOK, nil
}
