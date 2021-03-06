package apki

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/wenerme/apki/pkg/apki/apis"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func (s *IndexerServer) ServeWeb() error {
	c := restful.NewContainer()
	apis.MirrorResource{DB: s.DB}.RegisterTo(c)
	apis.PackageIndexResource{DB: s.DB}.RegisterTo(c)

	cors := restful.CrossOriginResourceSharing{
		AllowedDomains: []string{"localhost:3000", "alpine.ink"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
		Container:      c,
	}
	c.Filter(cors.Filter)
	c.Filter(c.OPTIONSFilter)
	restful.RegisterEntityAccessor(restful.MIME_JSON, entityJSONAccess{})

	r := mux.NewRouter()
	r.Use(recoveryMiddleware)
	r.Use(loggingMiddleware)

	r.PathPrefix("/api").Handler(http.StripPrefix("/api", c))
	r.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("PONG"))
	})
	logrus.Infof("serve %s", s.conf.Web.Addr)
	return http.ListenAndServe(s.conf.Web.Addr, r)
}

// default to jsoniter
type entityJSONAccess struct {
	ContentType string
}

func (e entityJSONAccess) Read(req *restful.Request, v interface{}) error {
	decoder := jsoniter.NewDecoder(req.Request.Body)
	decoder.UseNumber()
	return decoder.Decode(v)
}

func (e entityJSONAccess) Write(resp *restful.Response, status int, v interface{}) error {
	return e.writeJSON(resp, status, e.ContentType, v)
}

func (e entityJSONAccess) writeJSON(resp *restful.Response, status int, contentType string, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}
	// NOTE can not check is pretty
	if true {
		// pretty output must be created and written explicitly
		output, err := jsoniter.MarshalIndent(v, "", " ")
		if err != nil {
			return err
		}
		resp.Header().Set(restful.HEADER_ContentType, contentType)
		resp.WriteHeader(status)
		_, err = resp.Write(output)
		return err
	}
	// not-so-pretty
	resp.Header().Set(restful.HEADER_ContentType, contentType)
	resp.WriteHeader(status)
	return jsoniter.NewEncoder(resp).Encode(v)
}
