/* Server file implements a Server in order to make the external HTTP calls at the NASA's APIs */

package server

import (
	"littleExplorer/apod"
	"log"
	"net/http"
)

// Server setup and go online the server.
func Server() {
	log.Printf("Server() is running in a Go routine.")

	/* Create a muxer */
	var mux *http.ServeMux   // The HTTP request router (a.k.a., the muxer)
	mux = http.NewServeMux() // Instanciate the router.

	/* Create the Handlers */
	// 1st. A Re-Direct Handler to the main NASA API's.
	// Method 1: Use a pre-existed http func to return a Handler object.
	var apodNasaRedirect http.Handler // declare a handler var.
	apodNasaRedirect = http.RedirectHandler("https://api.nasa.gov/",
		http.StatusTemporaryRedirect) // assign the redirect http. function.

	// 2nd. A Handler to load the default Today's apod.
	// Method 2: Use a struct type + an http.ServeHTTP() method to Handle a route.
	// (the ServeHTTP method is located in the apod package. )
	var apodToday *apod.RespData // declare an *apod.RespData type var.

	// 3rd. A Handler to accept a date as a query parameter.
	// Method 3: Use a regular function with the apropriate signature and converte
	// the Handler with the use of http.HandlerFunc type.
	var apodWithDate http.HandlerFunc              // <- it's a type NOT a function.
	apodWithDate = http.HandlerFunc(apod.WithDate) // <- func (as value) to be converted.

	/* Register the Handler(s) into the muxer */
	mux.Handle("/nasaapis", apodNasaRedirect) // re-direction to NASA API's.
	mux.Handle("/apod", apodToday)            // the "today" APOD request.
	mux.Handle("/apod/", apodWithDate)        // APOD request with date.

	/* Setup and run a Server */
	log.Printf("Listening to localhost:3000...\n________________________________________________________________________\n\n")
	http.ListenAndServe("localhost:3000", mux)
}
