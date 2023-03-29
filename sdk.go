package unicornsdk

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/spf13/viper"
	"sync"
	"time"
)

type APIError struct {
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %s", e.Detail)
}

type PlatForm string

const (
	WINDOWS PlatForm = "WINDOWS"
	ANDROID PlatForm = "ANDROID"
	IOS     PlatForm = "IOS"
	OSX     PlatForm = "OSX"
)

func (self PlatForm) String() string {
	return string(self)
}

var instance *UnicornSdk
var once sync.Once
var client = req.C()

func GetSdkInstance() *UnicornSdk {
	once.Do(func() {
		instance = &UnicornSdk{}
		v := viper.New()

		baseURL := "https://us.unicorn-bot.com"
		defaultTimeout := 30 * time.Second

		// default config
		instance.v = v

		//client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

		client.SetBaseURL(baseURL).
			SetCommonError(&APIError{}).
			SetTimeout(defaultTimeout).
			SetCookieJar(nil).
			OnBeforeRequest(func(client *req.Client, req *req.Request) error {
				req.SetBearerAuthToken(GetSdkInstance().GetAcessToken())
				return nil
			}).
			OnAfterResponse(func(client *req.Client, resp *req.Response) error {
				if resp.Err != nil {
					// resp.Err represents the underlying error, e.g. network error, or unmarshal error (SetResult or SetError was invoked before).
					// Append dump content to original underlying error to help troubleshoot if request has been sent.
					if dump := resp.Dump(); dump != "" {
						resp.Err = fmt.Errorf("%s\nraw content:\n%s", resp.Err.Error(), resp.Dump())
					}
					return nil // Skip the following logic if there is an underlying error.
				}
				// Return a human-readable error if server api returned an error message.
				if err, ok := resp.Error().(*APIError); ok {
					if err.Error() == "Not authenticated" {
						resp.Err = NotAuthenticated{err.Detail}
					} else {
						resp.Err = err
					}
					return nil
				}
				// Corner case: neither an error response nor a success response (e.g. status code < 200),
				// dump content to help troubleshoot.
				if !resp.IsSuccess() {
					resp.Err = fmt.Errorf("bad response, raw content:\n%s", resp.Dump())
					return nil
				}
				return nil
			})

	})
	return instance
}

type UnicornSdk struct {
	v *viper.Viper
}

func (self *UnicornSdk) Auth(access_token string) *UnicornSdk {
	self.v.Set("access_token", access_token)
	return self
}

func (self *UnicornSdk) GetAcessToken() string {
	return self.v.GetString("access_token")
}

func (self *UnicornSdk) SetProxysForSdk(sdk_proxy string) *UnicornSdk {
	self.v.Set("sdk_proxy", sdk_proxy)
	client.SetProxyURL(sdk_proxy)
	return self
}

func (self *UnicornSdk) SetApiUrl(apiUrl string) *UnicornSdk {
	self.v.Set("api_url", apiUrl)
	client.SetBaseURL(apiUrl)
	return self
}

func (self *UnicornSdk) SetDebug(enable bool) *UnicornSdk {
	if enable {
		client.EnableDebugLog()
		client.EnableDumpAll()
	} else {
		client.DisableDebugLog()
		client.DisableDumpAll()
	}
	return self
}

func (self *UnicornSdk) SetTimeout(d time.Duration) *UnicornSdk {
	client.SetTimeout(d)
	return self
}

func (self *UnicornSdk) SetRootCertFromString(pemContent string) *UnicornSdk {
	client.SetRootCertFromString(pemContent)
	return self
}

func (self *UnicornSdk) CreateDeviceSession() *DeviceSession {
	deviceSession := &DeviceSession{}
	deviceSession.bundle = viper.New()
	return deviceSession
}
