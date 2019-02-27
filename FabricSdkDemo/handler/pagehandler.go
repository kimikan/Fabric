package handler

import (
	"FabricSdkDemo/sdk"
	"FabricSdkDemo/tools"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type HandlerContext struct {
	SdkKnife *sdk.SdkKnife
}

func (p *HandlerContext) IndexHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("template/index.tmpl")
	if err != nil {
		writeInfoToClient(w, "failed", err.Error())
		return
	}
	w.Write(content)
}


func (p *HandlerContext) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		content, err := ioutil.ReadFile("template/enroll.tmpl")
		if err != nil {
			writeInfoToClient(w, "failed", err.Error())
			return
		}
		w.Write(content)
	} else {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		org := r.PostFormValue("org")
		higher := r.PostFormValue("higher")
		if len(username) <= 0 {
			writeInfoToClient(w, "failed", "Username empty")
			return
		}
		if len(password) <= 0 {
			writeInfoToClient(w, "failed", "Password empty")
			return
		}
		if len(org) <= 0 {
			writeInfoToClient(w, "failed", "Org empty")
			return
		}
		if len(higher) <= 0 {
			writeInfoToClient(w, "failed", "Higher level empty")
			return
		}
		fmt.Println(username, password, org, higher)

	}
}

func (p *HandlerContext) AboutHandler(w http.ResponseWriter, r *http.Request) {
	writeInfoToClientByTemplate(w, "About", "written by kimikan", "template/about.tmpl")
}

func (p *HandlerContext) InvokeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("template/invoke.tmpl")
	if err != nil {
		writeInfoToClient(w, "failed", err.Error())
		return
	}
	w.Write(content)
}

func (p *HandlerContext) ChannelsHandler(w http.ResponseWriter, r *http.Request) {
	type channel struct {
		Index       int
		Name        string
		Description string
	}
	chans := []channel{
		channel{
			Index:       0,
			Name:        "ambrchannel",
			Description: "default channel",
		},
	}

	content, err := ioutil.ReadFile("template/channels.tmpl")
	t, e := template.New("webpage").Parse(string(content))
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}
	err = t.Execute(w, chans)
	if err != nil {
		writeInfoToClient(w, "Failed", err.Error())
	}
	/*
		out, e2 := exec.Command("bash", "-c", "./channels.sh").Output()
		if e2 != nil {
			fmt.Println(e2)
			writeInfoToClient(w, "failed", e2.Error())
		} else {
			fmt.Println(string(out))
			w.Write(content)
		}
	*/
}

func (p *HandlerContext) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		content, err := ioutil.ReadFile("template/transaction.tmpl")
		if err != nil {
			writeInfoToClient(w, "Failed", err.Error())
		} else {
			w.Write(content)
		}
	} else {
		id := r.PostFormValue("id")
		tx, e := p.SdkKnife.LedgerProvider.QueryTransaction(fab.TransactionID(id))
		if e != nil {
			writeInfoToClient(w, "Failed", e.Error())
		} else {
			writeInfoToClient(w, "Success", tx.String())
		}
	}

}

func (p *HandlerContext) TesterHandler(w http.ResponseWriter, r *http.Request) {
	result, err := ioutil.ReadFile("test.result")
	if err != nil {
		result = []byte("Empty")
	}

	content, err := ioutil.ReadFile("template/tester.tmpl")
	t, e := template.New("webpage").Parse(string(content))
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}

	err = t.Execute(w, string(result))
	if err != nil {
		writeInfoToClient(w, "Failed", err.Error())
	}
}

func JsonToObject(jsonStr string, obj interface{}) error {
	err := json.Unmarshal([]byte(jsonStr), obj)
	if err != nil {
		fmt.Println("Unmarshal failed, ", err)
		return err
	}
	return nil
}

func JsonToStrings(jsonStr string) ([]string, error) {
	strs := []string{}
	e := JsonToObject(jsonStr, &strs)
	if e != nil {
		return nil, e
	}

	return strs, nil
}

func ParseReqForm(r *http.Request) (string, string, []string, error) {
	ccid := r.PostFormValue("ccid")
	function := r.PostFormValue("function")
	json := r.PostFormValue("args")

	if len(ccid) <= 0 {
		return "", "", nil, errors.New("empty channel id")
	}

	if len(function) <= 0 {
		return "", "", nil, errors.New("empty function name")
	}

	args, e := JsonToStrings(json)
	if e != nil {
		fmt.Println("args error: ", args, e, json)
		return "", "", nil, e
	}

	return ccid, function, args, nil
}

func (p *HandlerContext) DoTesterHandler(w http.ResponseWriter, r *http.Request) {
	tstr := r.PostFormValue("threads")
	rstr := r.PostFormValue("rounds")
	fmt.Println(tstr, rstr)
	threads, e1 := strconv.Atoi(tstr)
	if e1 != nil {
		writeInfoToClient(w, "Failed", e1.Error())
		return
	}
	rounds, e2 := strconv.Atoi(rstr)
	if e2 != nil {
		writeInfoToClient(w, "Failed", e2.Error())
		return
	}
	//input values
	resp, e := tools.StressTest(threads, rounds)

	if e != nil {
		writeInfoToClient(w, "Failed", e.Error())
	} else {
		ioutil.WriteFile("test.result", []byte(resp), 0666)
		writeInfoToClient(w, "Stress testing result", resp)
	}
}

func (p *HandlerContext) DoInvokeHandler(w http.ResponseWriter, r *http.Request) {
	//input values
	ccid, function, args, e := ParseReqForm(r)

	if e != nil {
		writeInfoToClient(w, "Failed", e.Error())
	} else {
		resp, e2 := p.SdkKnife.Invoke(ccid, function, args)
		if e2 != nil {
			writeInfoToClient(w, "Failed", e2.Error())
		} else {
			writeInfoToClient(w, "Invoke success", fmt.Sprintf("%s\r\n\r\n %s", resp.TransactionID, string(resp.Payload)))
		}
	}
}

func (p *HandlerContext) DoQueryHandler(w http.ResponseWriter, r *http.Request) {
	//input values
	ccid, function, args, e := ParseReqForm(r)

	if e != nil {
		writeInfoToClient(w, "Failed", e.Error())
	} else {
		resp, e2 := p.SdkKnife.Query(ccid, function, args)
		if e2 != nil {
			writeInfoToClient(w, "Failed", e2.Error())
		} else {
			writeInfoToClient(w, "Query success", fmt.Sprint(string(resp.Payload)))
		}
	}
}

func (p *HandlerContext) QueryHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("template/query.tmpl")
	if err != nil {
		writeInfoToClient(w, "failed", err.Error())
		return
	}
	w.Write(content)
}

func writeInfoToClientByTemplate(w http.ResponseWriter, subject string, detail string, file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	type info struct {
		Subject string
		Detail  string
	}
	arg := &info{
		Subject: subject,
		Detail:  detail,
	}
	t, e := template.New("webpage").Parse(string(content))
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}

	err = t.Execute(w, arg)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func writeInfoToClient(w http.ResponseWriter, subject string, detail string) {
	writeInfoToClientByTemplate(w, subject, detail, "template/info.tmpl")
}
