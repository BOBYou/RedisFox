package main

import (
	"fmonitor/conf"
	"os"
	"flag"
	"fmonitor/flog"
	"strconv"
	"os/signal"
	"syscall"
	"sync"
	"fmonitor/util"
	"fmonitor/process"
)

var cpath string
var config *conf.Config

//初始化
func init() {
	flag.StringVar(&cpath, "config", "./conf/redis-fox.yaml", "config path with yml format")
	flag.Parse()
	if cpath == "" {
		flog.Fatalf("config path not found")
		os.Exit(1)
	}
	c, err := conf.NewConfig(cpath)
	if err != nil {
		flog.Fatalf(err.Error())
		os.Exit(1)
	}
	config = c
	flog.Init(config.Logname, config.Logpath, config.Loglevel)
	//StorePid("")
}

//存储pid
/*func StorePid(path string) {
	pid := os.Getpid()
	if len(path) == 0 {
		path = "./run_pusher.pid"
	}

	fout, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		os.Exit(1)
	}
	defer fout.Close()
	fout.WriteString(strconv.Itoa(pid))
}*/

func main() {
	/*fmt.Println(dataprovider.NewProvider(config).SaveInfoCommand("127.0.0.1:6379",retMap))*/
	wg := new(sync.WaitGroup)
	closeCh := make(chan struct{})
	probe := util.NewProbe(wg,closeCh)
	defer func() {
		close(closeCh)
		wg.Wait()
	}()
	for _,v := range config.Servers {
		server := v["server"]
		port, err := strconv.Atoi(v["port"])
		conntype := v["conntype"]
		if err != nil {
			flog.Fatalf(err.Error())
			os.Exit(1)
		}
		var passport string
		if v["passport"] != "" {
			passport = v["passport"]
		}

		//开启info
		info,infoErr := process.RunInfo(server,conntype,passport,port,config,probe)
		if infoErr != nil {
			flog.Warnf(info.ServerId+" run info error "+infoErr.Error())
		}

		//开启monitor
		//code...

		wg.Add(1)
	}
	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTERM)
	<-exitChan
	flog.Infof("fmonitor shut down")
}

