package wsapi

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"github.com/buger/jsonparser"
	"encoding/json"
	"fmt"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize: 4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type ApiMethod func ([]byte) (interface{})

type Api struct {
	Port string
	methods map[string]ApiMethod
}

type WSResponse struct {
	Method string `json:"m"`
	Data interface{} `json:"data"`
}

func NewApi(Port string) *Api {
	return &Api{Port, make(map[string]ApiMethod)}
}

func (a *Api) SetApiMethods(arr interface{}) {
	a.methods = make(map[string]ApiMethod)
	for k, v := range arr.(map[string]interface{}) {
		a.methods[k] = v.(func ([]byte) (interface{}))
	}
}

func(api *Api) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			api.Invoke(c, r.Header.Get("Sec-WebSocket-Key"))
		}
	})

	go http.ListenAndServe(":" + api.Port, mux)
}

func (api *Api)Invoke(conn *websocket.Conn, t string) {
	mt, message, err := conn.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return
	}

	
	method, _ := jsonparser.GetString(message, "m")
	if m, ok := api.methods[method]; ok {
		data, _,_,err := jsonparser.Get(message, "data")

		var result interface{}

		if (err != nil) {
			result = m(nil)
		} else {
			result = m(data)
		}

		response, _ := json.Marshal(WSResponse{method, result})
		
		conn.WriteMessage(mt, response)
	}
}