package openapi

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
)

var PrefixUri = "/api-doc/"
var FilePath = "./apidoc/"

func HttpHandle(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI[len(r.RequestURI)-5:] == ".json" {
		specDoc, err := loads.Spec(FilePath + "swagger.json")
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		specDoc, err = specDoc.Expanded(&spec.ExpandOptions{
			SkipSchemas:         false,
			ContinueOnError:     true,
			AbsoluteCircularRef: true,
		})
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//#nosec
		_, _ = w.Write(b)
		return
	}
	mod := r.RequestURI[len(PrefixUri):]
	middleware.Redoc(middleware.RedocOpts{
		BasePath: PrefixUri,
		SpecURL:  path.Join(PrefixUri, "swagger.json"),
		Path:     mod,
	}, http.NotFoundHandler()).ServeHTTP(w, r)
}
