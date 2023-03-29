package unicornsdk

import (
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/unicorn-aio/unicornsdk-go/utils"
	"gopkg.in/yaml.v2"
	"net/http"
)

type DeviceSession struct {
	bundle *viper.Viper
}

type DeviceInfo map[string]interface{}

func (self *DeviceSession) SetSessionID(session_id string) *DeviceSession {
	self.bundle.Set("session_id", session_id)
	return self
}

func (self *DeviceSession) SetPlatForm(platform PlatForm) *DeviceSession {
	self.bundle.Set("platForm", platform)
	return self
}

func (self *DeviceSession) SetAcceptLanguage(accept_language string) *DeviceSession {
	self.bundle.Set("accept_language", accept_language)
	return self
}

func (self *DeviceSession) SetUserAgent(user_agent string) *DeviceSession {
	self.bundle.Set("user_agent", user_agent)
	return self
}

func (self *DeviceSession) SetFlavors(key string, value interface{}) *DeviceSession {
	v := self.bundle.Get("flavors")
	var flavors map[string]interface{}
	if v == nil {
		flavors = make(map[string]interface{})
	} else {
		flavors = utils.InterfaceToMapInterface(v)
	}
	flavors[key] = value
	self.bundle.Set("flavors", flavors)
	return self
}

func (self *DeviceSession) GetUserAgent() string {
	ua := self.bundle.GetString("device_info.user_agent")
	if ua != "" {
		return ua
	}

	ua = self.bundle.GetString("user_agent")
	return ua
}

func (self *DeviceSession) GetDeviceInfo() (deviceinfo DeviceInfo) {
	v := self.bundle.Get("device_info")
	if v != nil {
		ret, ok := v.(DeviceInfo)
		if ok {
			return ret
		}

		m, ok := v.(map[string]interface{})
		if ok {
			ret := DeviceInfo(m)
			return ret
		}
		return deviceinfo
	}
	return deviceinfo
}

func (self *DeviceSession) GetSessionData() string {
	v := self.bundle.GetString("XSESSIONDATA")
	return v
}

func (self *DeviceSession) SerializeToMap() map[string]interface{} {
	return self.bundle.AllSettings()
}

func (self *DeviceSession) DeserializeFromMap(bundle map[string]interface{}) error {
	err := self.bundle.MergeConfigMap(bundle)
	return err
}

func (self *DeviceSession) DeserializeFromString(bundle string) error {
	c := make(map[string]interface{})
	bs := []byte(bundle)
	err := yaml.Unmarshal(bs, &c)
	if err != nil {
		return err
	}
	err = self.DeserializeFromMap(c)
	return err
}

func (self *DeviceSession) SerializeToString() (string, error) {
	c := self.SerializeToMap()
	bs, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (self *DeviceSession) GetCookies() []*http.Cookie {
	c := &http.Cookie{
		Name:  "XSESSIONDATA",
		Value: self.GetSessionData(),
	}
	return []*http.Cookie{c}
}

func (self *DeviceSession) InitSession() error {
	url := "/api/session/init/"

	sessionID := self.bundle.GetString("session_id")
	if sessionID == "" {
		sessionID = uuid.NewV4().String()
	}

	platform := WINDOWS
	p := self.bundle.Get("platform")
	if p != nil {
		platform = p.(PlatForm)
	}

	accept_language := self.bundle.GetString("accept_language")
	if accept_language == "" {
		accept_language = "en-US,en;q=0.9"
	}

	user_agent := self.bundle.GetString("user_agent")

	params := map[string]interface{}{
		"sessionid":       sessionID,
		"platform":        platform,
		"accept_language": accept_language,
	}

	if user_agent != "" {
		params["ua"] = user_agent
	}

	flavors := self.bundle.Get("flavors")
	if flavors != nil {
		params["flavors"] = flavors
	}

	deviceinfo := make(DeviceInfo)

	req := client.R()
	resp, err := req.SetBodyJsonMarshal(params).
		SetResult(&deviceinfo).
		Post(url)

	if resp.IsSuccess() {
		for _, c := range resp.Cookies() {
			if c.Name == "XSESSIONDATA" {
				self.bundle.Set("XSESSIONDATA", c.Value)
			}
		}
		self.bundle.Set("device_info", deviceinfo)
	}
	return err
}

func (self *DeviceSession) KasadaApi() *KasadaApi {
	api := &KasadaApi{
		DeviceSession: self,
	}
	api.state = viper.New()
	return api
}
