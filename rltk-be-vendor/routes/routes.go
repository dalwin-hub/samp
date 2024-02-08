package routes

import (
	"net/http"
	c "rltk-be-vendor/controllers"

	"github.com/gorilla/mux"
)

// InitializeRoutes has all the API routes defined and mapped
// with their middlewares
// This func gets initialised while starting the MUX server
func InitializeRoutes(router *mux.Router) {

	//master routes
	router.HandleFunc("/vendor/create",
		SetMiddlewareJSON(c.CreateVendor)).Methods(http.MethodPost)

	router.HandleFunc("/vendor/home",
		SetMiddlewareJSON(c.GetVendors)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/quickview/{id}",
		SetMiddlewareJSON(c.GetVendor)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/delete/{id}",
		SetMiddlewareJSON(c.DeleteVendor)).Methods(http.MethodDelete)

	router.HandleFunc("/vendor/edit/{id}",
		SetMiddlewareJSON(c.UpdateVendor)).Methods(http.MethodPut)

	router.HandleFunc("/vendor/filter/gettechnologies",
		SetMiddlewareJSON(c.GetFilterTechnologies)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/filter/getavailabledetails",
		SetMiddlewareJSON(c.GetAvailableDetails)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/filter",
		SetMiddlewareJSON(c.FilterVendor)).Methods(http.MethodPost)

	router.HandleFunc("/vendor/filter/test",
		SetMiddlewareJSON(c.FilterVendorTest)).Methods(http.MethodPost)

	router.HandleFunc("/vendor/country",
		SetMiddlewareJSON(c.GetCountry)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/country/state",
		SetMiddlewareJSON(c.GetState)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/country/state/city",
		SetMiddlewareJSON(c.GetCity)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/countrycode",
		SetMiddlewareJSON(c.GetCountryCode)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/edit/get/{id}",
		SetMiddlewareJSON(c.EditGetVendor)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/share",
		SetMiddlewareJSON(c.ShareMail)).Methods(http.MethodPost)

	router.HandleFunc("/vendor/search",
		SetMiddlewareJSON(c.Search)).Methods(http.MethodPost)

	router.HandleFunc("/vendor/allvendornames",
		SetMiddlewareJSON(c.GetVendorNames)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/createcontact",
		SetMiddlewareJSON(c.CreateVendorContact)).Methods(http.MethodPut)

	router.HandleFunc("/vendor/allvendorcontacts",
		SetMiddlewareJSON(c.GetVendorContactNames)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/validatevendorname",
		SetMiddlewareJSON(c.VendorNameValidate)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/validatevendorcontactname",
		SetMiddlewareJSON(c.VendorContactValidate)).Methods(http.MethodGet)

	router.HandleFunc("/vendor/downloaddocument",
		SetMiddlewareJSON(c.GetDocumentLink)).Methods(http.MethodGet)

}
