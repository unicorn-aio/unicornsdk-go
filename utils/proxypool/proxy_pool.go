package proxypool

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/unicorn-aio/unicornsdk-go/utils"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	EXCLUSIVE = iota
	REUSABLE
)

type InsufficientProxyError struct {
	Errmsg string
}

func (self InsufficientProxyError) Error() string {
	return self.Errmsg
}

type Proxy struct {
	protocol string
	username string
	password string
	host     string
	port     int
}

func (self *Proxy) ToURI() string {
	uri := ""
	protocal := self.protocol
	uri += protocal + "://"
	if self.username != "" && self.password != "" {
		uri += self.username + ":" + self.password + "@"
	}
	uri += self.host
	if self.port != 0 {
		uri += ":" + strconv.Itoa(self.port)
	}
	return strings.TrimSpace(uri)
}

type ProxyPool struct {
	proxys []string
	// backup proxys
	proxys_org []string
	mode       int
	mu         sync.Mutex
}

func Parse_ip_line(proxy_str, default_protocal string) (p Proxy, err error) {
	if proxy_str == "" {
		return p, fmt.Errorf("proxy_str 不能为空！")
	}
	var protocol, user, pass, host string
	var port = 0

	ss := strings.Split(proxy_str, "//")
	if len(ss) > 1 {
		protocol = strings.Trim(ss[0], ":")
		proxy_str = ss[1]
	} else {
		protocol = default_protocal
	}

	if strings.Index(proxy_str, "@") != -1 {
		// 正常的格式
		ss = strings.Split(proxy_str, "@")
		user_passwd := strings.Split(ss[0], ":")
		user = user_passwd[0]
		pass = user_passwd[1]

		ip_port := strings.Split(ss[1], ":")
		host = ip_port[0]
		port, err = strconv.Atoi(ip_port[1])
		if err != nil {
			return p, err
		}
	} else {
		ss = strings.Split(proxy_str, ":")
		if len(ss) == 2 {
			// 只有ip和端口
			host = ss[0]
			port, err = strconv.Atoi(ss[1])
			if err != nil {
				return p, err
			}
		} else if len(ss) == 4 {
			// 市面上常见的格式，通过端口号来判断
			port, err = strconv.Atoi(ss[1])
			if err == nil {
				// ip, port, user, pass 的格式
				host = ss[0]
				user = ss[2]
				pass = ss[3]
			} else {
				// user:pass:ip:port 的格式
				user = ss[0]
				pass = ss[1]
				host = ss[2]
				port, err = strconv.Atoi(ss[3])
				if err != nil {
					return p, err
				}
			}
		} else {
			err = errors.New("invalid proxy format!")
			return
		}
	}
	return Proxy{
		protocol: protocol,
		username: user,
		password: pass,
		host:     host,
		port:     port,
	}, nil
}

func LoadProxyFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	proxys := make([]string, 0)

	scaner := bufio.NewScanner(file)
	scaner.Split(bufio.ScanLines)
	for scaner.Scan() {
		line := scaner.Text()
		proxys = append(proxys, line)
	}
	return proxys, nil
}

func (self *ProxyPool) GetRandomOne() (string, error) {
	if len(self.proxys) > 0 {
		if self.mode == EXCLUSIVE {
			return self.PopOne()
		} else {
			n := rand.Intn(len(self.proxys))
			p := self.proxys[n]
			return p, nil
		}
	}
	return "", InsufficientProxyError{"no available proxy!"}
}

func (self *ProxyPool) PopOne() (string, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if len(self.proxys) > 0 {
		n := rand.Intn(len(self.proxys))
		p := self.proxys[n]
		self.proxys = utils.ArrayDeleteItem(self.proxys, p)
		return p, nil
	} else {
		return "", InsufficientProxyError{"no available proxy!"}
	}
}

func (self *ProxyPool) SetProxys(proxys []string) {
	self.proxys = proxys
	self.proxys_org = make([]string, 0)
	if self.proxys != nil {
		for _, i := range self.proxys {
			self.proxys_org = append(self.proxys_org, i)
		}
	}
}

func (self *ProxyPool) IsValid(proxy string) bool {
	if len(self.proxys_org) > 0 {
		for _, v := range self.proxys_org {
			if v == proxy {
				return true
			}
		}
	}
	return false
}

func EnsureLegalProxyFormat(proxy string) (string, error) {
	proxyObj, err := Parse_ip_line(proxy, "http")
	if err != nil {
		return "", err
	}
	return proxyObj.ToURI(), nil
}
