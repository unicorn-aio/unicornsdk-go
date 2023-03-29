package unicornsdk

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"github.com/spf13/viper"
	"io"
	"time"
)

type CdSolver func(rst time.Time, serverts time.Time, now time.Time) (string, error)

type KasadaApi struct {
	DeviceSession *DeviceSession
	state         *viper.Viper
	LocalCdSolver CdSolver
}

func (self *KasadaApi) SetXKpsdkCT(ct string) *KasadaApi {
	self.state.Set("x_kpsdk_ct", ct)
	self.state.Set("x_kpsdk_ct_refresh_time", time.Now().Unix())
	return self
}

func (self *KasadaApi) GetXKpsdkCT() string {
	return self.state.GetString("x_kpsdk_ct")
}

func (self *KasadaApi) SetXKpsdkST(st int64) *KasadaApi {
	self.state.Set("x_kpsdk_st", st)
	return self
}

func (self *KasadaApi) SetRst(rst int64) *KasadaApi {
	self.state.Set("rst", rst)
	return self
}

func (self *KasadaApi) SetUseProxyExitIp(enable bool) *KasadaApi {
	self.state.Set("proxy_exit_ip", enable)
	return self
}

func (self *KasadaApi) MarkRst() *KasadaApi {
	rst := time.Now().UnixMilli()
	st := self.state.GetInt64("x_kpsdk_st")
	self.state.Set("rst", rst)
	self.state.Set("st_diff", rst-st)
	return self
}

type KpSdkResp struct {
	XKpsdkCt       string            `json:"x_kpsdk_ct,omitempty"`
	XKpsdkCd       string            `json:"x_kpsdk_cd,omitempty"`
	XKpsdkCd2      string            `json:"x_kpsdk_cd2,omitempty"`
	XKpsdkSt       int               `json:"x_kpsdk_st,omitempty"`
	XKpsdkFp       string            `json:"x_kpsdk_fp,omitempty"`
	Url            string            `json:"url,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	CompressMethod string            `json:"compress_method,omitempty"`
	TlBodyB64      string            `json:"tl_body_b64,omitempty"`
}

func (self *KpSdkResp) DecompressBody() ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(self.TlBodyB64)
	if err != nil {
		return nil, err
	}
	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}

	payload, err := io.ReadAll(reader)
	return payload, err
}

func (self *KasadaApi) SolveCt(
	ips_url string,
	timezone_info string,
	referrer string,
	ips_content []byte,
) (kpResp *KpSdkResp, err error) {
	param := map[string]string{
		"ips_url": ips_url,
	}
	if timezone_info != "" {
		param["timezone_info"] = timezone_info
	}
	if referrer != "" {
		param["referrer"] = referrer
	}

	if self.state.GetBool("proxy_exit_ip") {
		param["proxy_exit_ip"] = "true"
	}

	var gzip_ips_content bytes.Buffer
	w := gzip.NewWriter(&gzip_ips_content)
	_, err = w.Write([]byte(ips_content))
	if err != nil {
		return nil, err
	}
	w.Close()

	req := client.R()
	_, err = req.SetFileBytes("ips_js", "ips_js", gzip_ips_content.Bytes()).
		SetQueryParams(param).
		SetCookies(self.DeviceSession.GetCookies()...).
		SetResult(&kpResp).
		Post("/api/kpsdk/ips/")
	return kpResp, err
}

func (self *KasadaApi) SolveCd() (kpKesp *KpSdkResp, err error) {
	return self.SolveCdWithNow(time.Time{})
}

func (self *KasadaApi) cdLongToCd(cd string) (string, error) {
	solve := make(map[string]interface{})
	err := json.Unmarshal([]byte(cd), &solve)
	if err != nil {
		return "", err
	}
	cdShort := map[string]interface{}{
		"workTime": solve["workTime"],
		"id":       solve["id"],
		"answers":  solve["answers"],
	}
	data, err := json.Marshal(cdShort)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (self *KasadaApi) SolveCdWithNow(now time.Time) (kpKesp *KpSdkResp, err error) {
	if self.LocalCdSolver != nil {
		rst := time.UnixMilli(self.state.GetInt64("rst"))
		st := time.UnixMilli(self.state.GetInt64("x_kpsdk_st"))
		var solve string
		solve, err = self.LocalCdSolver(rst, st, now)
		if err == nil {
			cdShort, innerErr := self.cdLongToCd(solve)
			if innerErr == nil {
				kpKesp = &KpSdkResp{
					XKpsdkCt:  self.GetXKpsdkCT(),
					XKpsdkCd:  cdShort,
					XKpsdkCd2: solve,
				}
				return
			}
		}
	}

	param := map[string]interface{}{
		"x_kpsdk_ct": self.state.GetString("x_kpsdk_ct"),
		"x_kpsdk_cr": true,
		"x_kpsdk_st": self.state.GetInt64("x_kpsdk_st"),
		"st_diff":    self.state.GetInt64("st_diff"),
		"rst":        self.state.GetInt64("rst"),
	}

	if !now.IsZero() {
		param["now_ms"] = now.UnixMilli()
	}

	req := client.R()
	req.SetBodyJsonMarshal(param)
	_, err = req.
		SetBodyJsonMarshal(param).
		SetCookies(self.DeviceSession.GetCookies()...).
		SetResult(&kpKesp).
		Post("/api/kpsdk/answer/")
	return kpKesp, err
}

func (self *KasadaApi) SerializeToMap() map[string]interface{} {
	return self.state.AllSettings()
}

func (self *KasadaApi) DeserializeFromMap(state map[string]interface{}) error {
	return self.state.MergeConfigMap(state)
}

func (self *KasadaApi) IsExpired() bool {
	if self.state.GetInt64("x_kpsdk_st") > 0 && self.state.GetString("x_kpsdk_ct") != "" && self.state.GetInt64("rst") > 0 && time.Now().Unix()-self.state.GetInt64("x_kpsdk_ct_refresh_time") <= 3600*22 {
		return false
	}
	return true
}
