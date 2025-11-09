package service

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"spider-go/internal/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Jwclogin(username string, jwcpassword string) string {
	// 1️ GET 请求获取登录页面及表单隐藏字段
	res, err := http.Get(utils.Jwc_url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("连接 WebVPN 失败: %v", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	lt := doc.Find("input[name='lt']").AttrOr("value", "")
	dllt := doc.Find("input[name='dllt']").AttrOr("value", "")
	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	eventID := doc.Find("input[name='_eventId']").AttrOr("value", "")
	salt := doc.Find("input[id='pwdDefaultEncryptSalt']").AttrOr("value", "")
	rmShown := doc.Find("input[name='rmShown']").AttrOr("value", "")

	//need, _ := http.Get(utils.Captcha_url + "username=" + username + "&pwdEncrypt2=pwdEncryptSalt" + "&_=" + strconv.FormatInt(time.Now().Unix(), 13))

	fmt.Println("lt:", lt, "dllt:", dllt, "execution:", execution, "eventID:", eventID, "salt:", salt)

	// 2️ 加密密码
	encryptedPwd := utils.JsCrypto(jwcpassword, salt)
	log.Println("encryptedPwd:", encryptedPwd)

	// 3️ 构造 URL 编码的表单数据
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", encryptedPwd)
	form.Set("lt", lt)
	form.Set("dllt", dllt)
	form.Set("execution", execution)
	form.Set("_eventId", eventID)
	form.Set("rmShown", rmShown)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 返回错误，阻止任何重定向，之后在这里直接处理redirect
			println(req.URL.String())
			return http.ErrUseLastResponse
		},
	}

	// 4️ 创建 POST 请求
	request, err := http.NewRequest("POST", utils.Jwc_url, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	// 添加 cookies
	for _, cookie := range res.Cookies() {
		request.AddCookie(cookie)
	}

	// 设置 headers
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36 Edg/142.0.0.0")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// 5️ 发送请求
	resp, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println("Response Status Code:", resp.StatusCode)

	return strconv.Itoa(resp.StatusCode)
}
