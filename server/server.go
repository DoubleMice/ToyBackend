package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"./Mariadb"
	"strconv"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

/** response json
{
	"status" 		: "invalid" | "oldUser" | "newUser" | "unknow"
	"openId" 		: openid
	"targetOpenId"	: pair openid
}
status:
	invalid : wechar server response invalid
	oldUser : openid exist
	newUser : openid not exist
	unknow	: other situation
openid:
	get from wechat server
targetOpenId:
	for oldUser,which has made a pair

**/
func loginHandler(w http.ResponseWriter, r *http.Request) {

	type resType struct {
		Status string `json:"status"`
		OpenId string `json:"openId"`
		TargetOpenId string `json:"targetOpenId"`
	}
	res2Front := resType{"","",""}
	r.ParseForm()
	js_code := r.FormValue("code")
	if js_code == "" {
		return
	}
	requestUrl := fmt.Sprintf(mariadb.Code2SessionAPI,mariadb.AppId,mariadb.AppSecret,js_code,"authorization_code")
	fmt.Println(requestUrl)
	resp,err := http.Get(requestUrl)
	if err!=nil {
		fmt.Println(err)
		return
	} 
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var wxRes interface{}
	err = json.Unmarshal(body,&wxRes)
	if err != nil {
		fmt.Println(err)
		return
	}
	resMap := wxRes.(map[string]interface{})
	if resMap["errcode"] != nil {
		// request wechat server fail
		errmsg := fmt.Sprintf("%s",resMap["errmsg"])
		println("errmsg:"+errmsg)
		res2Front.Status = "invalid"
	} else {
		fmt.Println(resMap)
		// request wechat server success
		openId := fmt.Sprintf("%s",resMap["openid"])
		if mariadb.QueryIfUserExist(openId) == true {
			// has record
			res2Front.Status = "oldUser"
			res2Front.OpenId = openId
			res2Front.TargetOpenId = mariadb.QueryPairGetTarget(openId)
		} else {
			// no record
			res2Front.Status = "newUser"
			res2Front.OpenId = openId
		}
	}
	res2FrontJsonfy,err := json.Marshal(res2Front)
	println(string(res2FrontJsonfy))
	w.Write(res2FrontJsonfy)
}


/** response json

{
	"status" 		: "paramError" | "unknow" | "unique" | "ok"
	"targetOpenId" 	: pair openid
}
status:
	paramError 	: request params error
	unknow 		: convert or query error
	unique 		: only 1 user , but insert user success
	registerOk 	: make a pair success
targetOpenId	: if status == "ok" , means had build pair relationship and 
				  could get a target openId to build a ws channal
**/
func quizHandler(w http.ResponseWriter, r *http.Request) {
	type resType struct {
		Status string `json:"status"`
		TargetOpenId string `json:"targetOpenId"`
	}
	var res2Front resType

	r.ParseForm()
	answerStr := r.FormValue("answer")
	openId := r.FormValue("openId")
	if (answerStr == "" || openId == "") {
		res2Front.Status = "paramError"
		res2FrontJsonfy,_ := json.Marshal(res2Front)
		w.Write(res2FrontJsonfy)
		return
	}
	answer, err := strconv.ParseInt(answerStr, 10, 64)
	if err != nil {
		println("convert string to int64 fail")
		res2Front.Status = "unknow"
		res2FrontJsonfy,_ := json.Marshal(res2Front)
		w.Write(res2FrontJsonfy)
		return
	}
	if mariadb.QueryIfUserExist(openId) == false {
		mariadb.InsertUser(openId,answer)
		res2Front.Status = mariadb.MakePair(openId,answer)
	} else {
		res2Front.Status = mariadb.MakeNewPair(openId,answer)
	}
	res2Front.TargetOpenId = mariadb.QueryPairGetTarget(openId)
	res2FrontJsonfy,err := json.Marshal(res2Front)
	println(string(res2FrontJsonfy))
	w.Write(res2FrontJsonfy)
}



var chatPool = make(map[string]*Hub)
func main(){
	http.HandleFunc("/", myHandler)		//	设置访问路由
	http.HandleFunc("/login",loginHandler)
	http.HandleFunc("/quiz",quizHandler)
	http.HandleFunc("/chat",func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		openId := r.FormValue("openId")
		targetOpenId := r.FormValue("targetOpenId")
		println(openId)
		println(targetOpenId)
		if (openId == ""||targetOpenId == "") {
			// println("error param")
			// return
		}
		if (chatPool[openId] == nil || chatPool[targetOpenId] == nil) {
			hub := newHub()
			chatPool[openId] = hub
			chatPool[targetOpenId] = hub
			go hub.run()
			serveWs(hub, w, r)
		} else {
			serveWs(chatPool[openId],w,r)
		}
	})
	// fmt.Println(http.ListenAndServe(":8080", nil))
	fmt.Println(http.ListenAndServeTLS(":443","ssl.crt","ssl.key",nil))
}