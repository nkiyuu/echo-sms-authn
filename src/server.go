package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

type Message struct {
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Message   string `json:"message"`
}

type Respons struct {
	ID string `json:"id"`
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Hello(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}

func Ok(c echo.Context) error {
	return c.Render(http.StatusOK, "ok", nil)
}

func Ng(c echo.Context) error {
	return c.Render(http.StatusOK, "ng", nil)
}

func Sms(c echo.Context) error {
	url := "https://api.cmtelecom.com/v1.0/otp/generate"

	msg := Message{
		Recipient: "{{{Pone Number}}}",
		Sender:    "test",
		Message:   "認証用の番号は {code} です",
	}

	postData, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(postData))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CM-ProductToken", "{{{Your Token}}}")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	var resBody Respons
	b, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(b, &resBody)

	return c.Render(http.StatusOK, "sms", resBody.ID)
}

type authMessage struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type authMessageResponse struct {
	Valid bool `json:"valid"`
}

func Auth(c echo.Context) error {
	url := "https://api.cmtelecom.com/v1.0/otp/verify"
	id := c.FormValue("id")
	code := c.FormValue("code")

	msg := authMessage{
		ID:   id,
		Code: code,
	}

	log.Println(msg)

	postData, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(postData))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CM-ProductToken", "{{{Your Token}}}")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	var resBody authMessageResponse
	b, _ := ioutil.ReadAll(res.Body)
	log.Println(string(b))
	json.Unmarshal(b, &resBody)

	log.Println(resBody.Valid)
	isValid := resBody.Valid

	if isValid {
		return c.Render(http.StatusOK, "ok", nil)
	} else {
		return c.Render(http.StatusOK, "ng", nil)
	}
}

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("src/views/*.html")),
	}
	e := echo.New()
	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/hello", Hello)
	e.GET("/sms", Sms)
	e.POST("/sms", Auth)

	e.Logger.Fatal(e.Start(":1323"))
}
