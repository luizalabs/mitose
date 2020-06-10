package admin

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/controller"
	"github.com/luizalabs/mitose/k8s"
)

type Handler struct {
	tmpl *template.Template
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.Get(w, r)
	} else {
		h.Post(w, r)
	}
}

func (h *Handler) Get(w http.ResponseWriter, _ *http.Request) {
	cm, err := k8s.GetConfigMapData("mitose", "config") // FIXME: unify those names
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := make(map[string]*config.Config)
	for k, v := range cm {
		conf := new(config.Config)
		if err := json.Unmarshal([]byte(v), conf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data[k] = conf
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	confName := r.FormValue("name")
	cm, err := k8s.GetConfigMapData("mitose", "config")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var min, max int
	if min, err = strconv.Atoi(r.FormValue("min")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if max, err = strconv.Atoi(r.FormValue("max")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.FormValue("type") {
	case "sqs":
		sqsConf := new(controller.SQSControlerConfig)
		if err := json.Unmarshal([]byte(cm[confName]), sqsConf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sqsConf.Min = min
		sqsConf.Max = max
		b, err := json.Marshal(sqsConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cm[confName] = string(b)
	case "pubsub":
		pubsubConf := new(controller.PubSubControlerConfig)
		if err := json.Unmarshal([]byte(cm[confName]), pubsubConf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pubsubConf.Min = min
		pubsubConf.Max = max
		b, err := json.Marshal(pubsubConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cm[confName] = string(b)
	case "rabbitmq":
		rabbitmqConf := new(controller.RabbitMQControlerConfig)
		if err := json.Unmarshal([]byte(cm[confName]), rabbitmqConf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rabbitmqConf.Min = min
		rabbitmqConf.Max = max
		b, err := json.Marshal(rabbitmqConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cm[confName] = string(b)
	}
	if err := k8s.UpdateConfigMap("mitose", cm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
}

func NewHandler() (http.Handler, error) {
	tmpl, err := template.ParseFiles("admin/index.html")
	if err != nil {
		return nil, err
	}
	return &Handler{tmpl}, nil
}
