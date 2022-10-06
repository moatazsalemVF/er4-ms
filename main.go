package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moatazsalemVF/er4-ms/er4tools"
	"github.com/moatazsalemVF/ms-template/utils"
)

var cmds = []string{}

func er4Write(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	body := strings.TrimSpace(string(bodyBytes))
	cmds = append(cmds, body)
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Received"))
}

func er4Read(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	body := strings.TrimSpace(string(bodyBytes))
	result := er4tools.Read(body)
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func lbState(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(er4tools.GetLBStatus())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func linkState(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(er4tools.GetLinkStatus())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func setLBState(w http.ResponseWriter, r *http.Request) {
	var exec er4tools.Executor
	err := json.NewDecoder(r.Body).Decode(&exec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(er4tools.SetLBStatus(exec)))
}

func processCMDs() {
	for {
		if len(cmds) > 0 {
			cmd := cmds[0]
			cmds = append(cmds[:0], cmds[1:]...)
			processCMD(cmd)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func processCMD(cmd string) {
	splt := strings.Split(cmd, "|")
	if splt[0] == "0" {
		er4tools.Disable(splt[1])
	} else {
		er4tools.Enable(splt[1])
	}
}

func main() {
	utils.Initialize()
	addControllers()
	go processCMDs()
	utils.Router.Run(utils.Conf.Server.Address + ":" + fmt.Sprint(utils.Conf.Server.Port))
}

func addControllers() {

	utils.Router.POST("/exec/write", func(c *gin.Context) {
		er4Write(c.Writer, c.Request)
	})

	utils.Router.POST("/exec/read", func(c *gin.Context) {
		er4Read(c.Writer, c.Request)
	})

	utils.Router.POST("/exec/command", func(c *gin.Context) {
		setLBState(c.Writer, c.Request)
	})

	utils.Router.GET("/er/lbstat", func(c *gin.Context) {
		lbState(c.Writer, c.Request)
	})

	utils.Router.GET("/er/linkstat", func(c *gin.Context) {
		linkState(c.Writer, c.Request)
	})

}
