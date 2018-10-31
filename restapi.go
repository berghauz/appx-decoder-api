package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// runAPIServer run app server
func (ctx *Context) runAPIServer() {
	router := mux.NewRouter()

	router.HandleFunc("/deveui/{deveui}/{payload}", decodeByDeveui).Methods("GET")
	router.HandleFunc("/model/{model}/{payload}", ctx.decodeByModel).Methods("GET")
	router.HandleFunc("/info", ctx.listAllDrivers).Methods("GET")
	router.HandleFunc("/info/{devtype}", ctx.getDriverInfo).Methods("GET")

	logger.Fatalln(http.ListenAndServe(ctx.ListenHost, router))
}

func decodeByDeveui(w http.ResponseWriter, r *http.Request) {

	//var data interface{}
	defer r.Body.Close()

	//body, _ := ioutil.ReadAll(r.Body)
	//json.Unmarshal(body, &data)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", mux.Vars(r)["deveui"])
}

// !!! change w.WriteHeader in all api calls !!!
func (ctx *Context) decodeByModel(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)

	if decoder, ok := ctx.DecodingPlugins[mux.Vars(r)["model"]]; ok {
		payload, err := decoder.Decode(mux.Vars(r)["payload"])
		if err != nil {
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
		} else {
			bstr, err := json.Marshal(payload)
			if err != nil {
				fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
				return
			}
			fmt.Fprintln(w, string(bstr))
		}
	} else {
		fmt.Fprintln(w, "{\"error\":\"unknown decoder model\"}")
	}

	//fmt.Fprintln(w, "{\"error\":\"malfunctioned request\"}")
}

func (ctx *Context) listAllDrivers(w http.ResponseWriter, r *http.Request) {

	var allDecoders []string
	//allDecoders := make(map[string]string)
	defer r.Body.Close()

	for decoderType := range ctx.DecodingPlugins {
		allDecoders = append(allDecoders, decoderType)
		//allDecoders[decoderType] = ctx.DecodingPlugins[decoderType].Version
	}

	data, err := json.Marshal(allDecoders)
	if err != nil {
		logger.Errorf("error:", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", `{"error": "internal error"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v", string(data))
}

func (ctx *Context) getDriverInfo(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if decoder, ok := ctx.DecodingPlugins[mux.Vars(r)["devtype"]]; ok {
		data, err := json.Marshal(decoder)
		if err != nil {
			logger.Errorf("error:", err)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", `{"error": "internal error"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", string(data))
	} else {
		//logger.Infoln(mux.Vars(r)["devtype"])
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", `{"error": "decoder not found"}`)
	}
}
