package er4tools

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/helloyi/go-sshclient"
)

const router string = ""
const sshPort string = ""
const user string = ""
const pass string = ""

func Disable(rule string) {
	fmt.Println("disable: " + rule)
	cookies := login()
	disableER(rule, cookies)
	logout(cookies)
}

func Enable(rule string) {
	fmt.Println("enable: " + rule)
	cookies := login()
	enableER(rule, cookies)
	logout(cookies)
}

func login() []*http.Cookie {
	apiUrl := "https://" + router
	resource := "/"
	data := url.Values{}
	data.Set("username", user)
	data.Set("password", pass)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	return resp.Cookies()
}

func enableER(rule string, cookies []*http.Cookie) {
	apiUrl := "https://" + router
	resource := "/api/edge/set.json"

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	data := "{\"firewall\":{\"name\":{\"Block_Devices\":{\"rule\":{\"" + rule + "\":{\"disable\":null}}}}}}"

	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("X-CSRF-TOKEN", cookies[1].Value)

	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}

	_, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
}

func disableER(rule string, cookies []*http.Cookie) {
	apiUrl := "https://" + router
	resource := "/api/edge/batch.json"

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	data := "{\"DELETE\":{\"firewall\":{\"name\":{\"Block_Devices\":{\"rule\":{\"" + rule + "\":{\"disable\":null}}}}}}}"

	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("X-CSRF-TOKEN", cookies[1].Value)

	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}

	_, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
}

func logout(cookies []*http.Cookie) {
	apiUrl := "https://" + router
	resource := "/logout"

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	r, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	r.Header.Add("X-CSRF-TOKEN", cookies[1].Value)

	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}

	_, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
}

type LoadbalanceGroup struct {
	Name           string `json:"name"`
	BalanceLocal   string `json:"balanceLocal"`
	LockLocalDNS   string `json:"lockLocalDNS"`
	ConntrackFlush string `json:"conntrackFlush"`
	StickyBits     string `json:"stickyBits"`
	Eths           []Eth  `json:"interfaces"`
}

type Eth struct {
	Name        string `json:"name"`
	Reachable   string `json:"reachable"`
	Status      string `json:"status"`
	Gateway     string `json:"gateway"`
	Routetable  string `json:"routeable"`
	Weight      string `json:"weight"`
	FO_priority string `json:"foPriority"`
	WANOut      string `json:"wanOut"`
	WANIn       string `json:"wanIn"`
	LocalICMP   string `json:"localICMP"`
	LocalDNS    string `json:"localDNS"`
	LocalData   string `json:"localData"`
}

type Link struct {
	Name            string `json:"name"`
	AutoNegotiation string `json:"autoNegotiation"`
	Speed           string `json:"speed"`
	Duplex          string `json:"duplex"`
	LinkDetected    string `json:"linkdDetected"`
}

type Executor struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func GetLinkStatus() []Link {
	client, err := sshclient.DialWithPasswd(router+":"+sshPort, user, pass)
	if err != nil {
		log.Panic(err)
	}

	out, err := client.Cmd("/opt/vyatta/bin/vyatta-op-cmd-wrapper show interfaces ethernet physical").Output()
	if err != nil {
		log.Panic(err)
	}

	data := strings.Split(string(out), "\n")
	links := []Link{}
	linkIdx := -1
	allIdx := 0
	for _, datum := range data {
		val := strings.Replace(datum, " ", "", -1)

		if val == "" {
			continue
		}

		if allIdx == 0 {
			linkIdx++
			link := Link{}
			link.Name = "eth" + fmt.Sprint(linkIdx)
			links = append(links, link)
			allIdx++
		} else if allIdx == 1 {
			links[linkIdx].AutoNegotiation = strings.SplitN(val, ":", -1)[1]
			allIdx++
		} else if allIdx == 2 {
			links[linkIdx].Speed = strings.SplitN(val, ":", -1)[1]
			allIdx++
		} else if allIdx == 3 {
			links[linkIdx].Duplex = strings.SplitN(val, ":", -1)[1]
			allIdx++
		} else if allIdx == 4 {
			links[linkIdx].LinkDetected = strings.SplitN(val, ":", -1)[1]
			allIdx = 0
		}

	}

	defer client.Close()
	return links
}

func GetLBStatus() LoadbalanceGroup {
	client, err := sshclient.DialWithPasswd(router+":"+sshPort, user, pass)
	if err != nil {
		log.Panic(err)
	}

	out, err := client.Cmd("/opt/vyatta/bin/vyatta-op-cmd-wrapper show load-balance status").Output()
	if err != nil {
		log.Panic(err)
	}

	data := strings.Split(string(out), "\n")

	lbGroup := LoadbalanceGroup{}
	eths := []Eth{}
	ethIdx := -1

	for index, datum := range data {
		val := strings.Replace(datum, " ", "", -1)

		if index == 0 {
			lbGroup.Name = val
		} else if index == 1 {
			lbGroup.BalanceLocal = strings.SplitN(val, ":", -1)[1]
		} else if index == 2 {
			lbGroup.LockLocalDNS = strings.SplitN(val, ":", -1)[1]
		} else if index == 3 {
			lbGroup.ConntrackFlush = strings.SplitN(val, ":", -1)[1]
		} else if index == 4 {
			lbGroup.StickyBits = strings.SplitN(val, ":", -1)[1]
		} else if index == 6 || index == 20 {
			ethIdx++
			eth := Eth{}
			eth.Name = strings.SplitN(val, ":", -1)[1]
			eths = append(eths, eth)
		} else if index == 7 || index == 21 {
			eths[ethIdx].Reachable = strings.SplitN(val, ":", -1)[1]
		} else if index == 8 || index == 22 {
			eths[ethIdx].Status = strings.SplitN(val, ":", -1)[1]
		} else if index == 9 || index == 23 {
			eths[ethIdx].Gateway = strings.SplitN(val, ":", -1)[1]
		} else if index == 10 || index == 24 {
			eths[ethIdx].Routetable = strings.SplitN(val, ":", -1)[1]
		} else if index == 11 || index == 25 {
			eths[ethIdx].Weight = strings.Replace(strings.SplitN(val, ":", -1)[1], "%", "", -1)
		} else if index == 12 || index == 26 {
			eths[ethIdx].FO_priority = strings.SplitN(val, ":", -1)[1]
		} else if index == 14 || index == 28 {
			eths[ethIdx].WANOut = strings.SplitN(val, ":", -1)[1]
		} else if index == 15 || index == 29 {
			eths[ethIdx].WANIn = strings.SplitN(val, ":", -1)[1]
		} else if index == 16 || index == 30 {
			eths[ethIdx].LocalICMP = strings.SplitN(val, ":", -1)[1]
		} else if index == 17 || index == 31 {
			eths[ethIdx].LocalDNS = strings.SplitN(val, ":", -1)[1]
		} else if index == 18 || index == 32 {
			eths[ethIdx].LocalData = strings.SplitN(val, ":", -1)[1]
		}
	}

	lbGroup.Eths = eths

	defer client.Close()

	return lbGroup
}

func SetLBStatus(exec Executor) string {
	client, err := sshclient.DialWithPasswd(router+":"+sshPort, user, pass)
	if err != nil {
		log.Panic(err)
	}

	client.Cmd("/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper begin").Run()
	cmds := strings.Split(exec.Command, "|")
	for _, cmd := range cmds {
		client.Cmd("/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper " + cmd).Run()
	}
	client.Cmd("/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper commit").Run()
	client.Cmd("/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper save").Run()

	defer client.Close()
	return "done"
}

func Read(cmd string) string {
	client, err := sshclient.DialWithPasswd(router+":"+sshPort, user, pass)
	if err != nil {
		log.Panic(err)
	}

	out, err := client.Cmd("/opt/vyatta/bin/vyatta-op-cmd-wrapper " + cmd).Output()
	if err != nil {
		log.Panic(err)
	}

	defer client.Close()
	return string(out)
}
