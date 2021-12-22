package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/youshintop/apiserver/pkg/db"
	v "github.com/youshintop/apiserver/pkg/version"
	"github.com/youshintop/apiserver/router/middleware"
	"github.com/youshintop/log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/youshintop/apiserver/config"
	"github.com/youshintop/apiserver/router"
)

var (
	cfg     = pflag.StringP("config", "c", "", "app config file path.")
	version = pflag.BoolP("version", "v", false, "show version info.")
)

func main() {

	pflag.Parse()
	if *version {
		version2 := v.Get()
		marshalled, err := json.MarshalIndent(&version2, "", " ")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(marshalled))
		return
	}

	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	db.Database()
	opts := &log.Options{
		Level:        "debug",
		Format:       "console",
		EnableCaller: true,
		EnableColor:  false,
		OutputPaths:  []string{"stdout"},
	}
	_, err := log.New(opts)
	if err != nil {
		panic(err)
	}
	defer log.Flush()

	gin.SetMode(viper.GetString("mode"))

	g := gin.New()

	router.Load(g, middleware.RequestId())

	go func() {
		if err := pingServer(); err != nil {
			log.Fatalf("The router has no response, or it might took too long to start up.", log.Err(err))
		}
		log.Info("The router has been deployed successfully.")
	}()

	log.Infof("Start to listening the incoming requests on http address %s", viper.GetString("address"))
	log.Info(http.ListenAndServe(viper.GetString("address"), g).Error())
}

func pingServer() error {
	for i := 0; i < viper.GetInt("max_ping_count"); i++ {
		resp, err := http.Get("http://127.0.0.1:8080/sd/health")

		if err == nil && resp.StatusCode == 200 {
			return nil
		}

		log.Info("Waiting for the router, retry in 1 second.")
		time.Sleep(time.Second)
	}

	return errors.New("cannot connect to the router")
}
