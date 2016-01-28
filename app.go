package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

//{"data":{"count":0,"wodou":650,"getWodou":50},"message":null,"success":true}
type SignJsonResultData struct {
	Count, Wodou, GetWodou int64
}

type SignJsonResult struct {
	Success bool
	Message string
	Data    SignJsonResultData
}

type MemberDayData struct {
	code int64
}

type MemberDay struct {
	Success bool
	Message string
	Data    MemberDayData
}

//{"data":{"message":{"show":"0","chance":0,"info":"7M\u6d41\u91cf"}},"message":null,"success":true}
type RedPacketDataInfo struct {
	Show   string
	Chance int64
	Info   string
}
type RedPacketData struct {
	Message RedPacketDataInfo
}
type RedPacket struct {
	Data    RedPacketData
	Message string
	Success bool
}

func login(mobile, password string, cookieJar *cookiejar.Jar) (http.Client, error) {
	loginUrl := "http://17wo.cn/Login!process.action"
	log.Println("登录网址 ", loginUrl)

	postBody := url.Values{}
	postBody.Set("mobile", mobile)
	postBody.Add("password", password)
	postBody.Add("chkType", "on")
	postBody.Add("backurl", "")
	postBody.Add("backurl2", "")
	postBody.Add("chk", "")
	postBody.Add("loginType", "0")

	client := http.Client{
		Jar: cookieJar,
	}

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(postBody.Encode()))
	if err != nil {
		log.Println("some wrong :", err)
		return client, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; U; CPU iPhone OS 3_0 like Mac OS X; en-us) AppleWebKit/528.18 (KHTML, like Gecko) Version/4.0 Mobile/7A341 Safari/528.16")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return client, err
	}

	htmlStr := string(body)

	matched, _ := regexp.MatchString(`<p id="getAlert"[^>]*>(.*?)</p>|(错误信息)`, htmlStr)
	if resp.StatusCode != http.StatusOK || matched {
		return client, errors.New("登录失败,请检查你的手机号或密码是否正确")
	}

	log.Println("登录成功")

	matched, _ = regexp.MatchString(`已签到`, htmlStr)
	if matched {
		log.Println("**已签到过了,无需再签**")
	}

	return client, nil
}

func requestJson(fn func([]byte), client http.Client, cookieJar *cookiejar.Jar, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("request json failed. ", err)
		return
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	fn(body)

	log.Println("")
}

func signin(body []byte) {
	log.Println("====  签到 开始 ====")

	var m SignJsonResult
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("%+v", m)
	log.Println("====  签到 结束 ====")
}

func memberDay(body []byte) {
	log.Println("====  会员日开红包 开始 ====")

	var m MemberDay
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("%+v", m)
	log.Println("====  会员日开红包 结束 ====")
}

func flowRedPacket(body []byte) {
	log.Println("====  流量红包 开始 ====")

	var m RedPacket
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("%+v", m)
	log.Println("====  流量红包 结束 ====")
}

func main() {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	cookieJar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}

	mobile := flag.Int64("mobile", 0, "手机号")
	password := flag.String("password", "", "密码")

	flag.Parse()

	if *mobile == 0 || *password == "" {
		log.Println("请使用 auto17wo -h 查看帮助")
		return
	}

	log.Println("手机: ", *mobile)
	log.Println("密码: ", *password)

	client, err := login(strconv.FormatInt(*mobile, 10), *password, cookieJar)
	if err != nil {
		log.Println(err)
		return
	}

	requestJson(signin, client, cookieJar, "http://17wo.cn/SignIn!checkin.action?checkIn=true&rnd=4151")
	requestJson(memberDay, client, cookieJar, "http://17wo.cn/MemberDay!draw.action?_=1411834905292")
	requestJson(flowRedPacket, client, cookieJar, "http://17wo.cn/FlowRedPacket!LuckDraw.action?1238")

}
