/* APOD: Astronomy Picture of the Day */
// https://api.nasa.gov/planetary/apod?api_key="???"&date=today&hd=True

package apod

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Only 3 Globals. Please don't kill me, right?
var (
	qCounter  int                // tracks the number of queries.
	apodQuery Query              // declare a var of type Query.
	apodTmpl  *template.Template // the basic template for apod.
)

// Query struct holds the query parameters we use to pass into the query.
type Query struct {
	APIKey     string // The API key
	DateString string // The date in STRING in YYYY-MM-DD of the image to retrieve.
	HD         string // A string boolean to retrieve (OR not) the HD URL image.
}

// BuildQuery method returns the query to be run.
func (qr *Query) BuildQuery() string {
	var query string
	query = fmt.Sprintf("?api_key=%s&date=%s&hd=%s", qr.APIKey, qr.DateString, qr.HD)
	return query // return the query
}

// RespData is the response data containing all the info and links, returned from APOD. // It is used when we UnMarshall to struct.
type RespData struct {
	/* Possible fileds combination:
	[copyright(?) date explanation hdurl(?) media_type service_version title url]
	Note: copyright and hdurl are not present in every response. */
	Copyright      string `json:"copyright"`       // © copyright information.
	Date           string `json:"date"`            // the date we want the picture.
	Explanation    string `json:"explanation"`     // a colophon about the image.
	HDUrl          string `json:"hdurl"`           // HD availability.
	MediaType      string `json:"media_type"`      // image, OR video?
	ServiceVersion string `json:"service_version"` // I don't know!!!
	Title          string `json:"title"`           // The title.
	URL            string `jcon:"url"`             // the URl of the image.
}

// The classic ServeHTTP method to run. Used ONLY for todays routing.
func (rd *RespData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		resp     *http.Response // create an *http.Response var.
		respBody []byte         // the response body.
		err      error          // handling the errors.
		path     string         // the server URL route.
	)
	path = r.URL.Path                               // the full route URL path.
	apodQuery.DateString = getTodaysDate()          // default: Today's date!
	log.Printf("Server Route Path: %v\n", path)     // print the route path.
	resp = makeRequest(apodQuery)                   // make the request and get response.
	respBody = processTheResponse(resp)             // process the response.
	err = json.Unmarshal(respBody, &rd)             // unmarshal to the struct.
	checkError("Error unmarshaling JSON data", err) // check for error.

	err = apodTmpl.ExecuteTemplate(w, "apod", rd) // execute: write to w.
	checkError("Error executing template", err)   // check for error.
}

const (
	// URL The URL to call for APOD API.
	URL string = "https://api.nasa.gov/planetary/apod"
)

func init() {
	fmt.Println()
	log.Println("__INIT__: APOD Initiallization. init() is running.")
	var err error // handling the errors.

	/* Authenticate to API & create the "Today" query */
	apodQuery.APIKey = GetApodAPIKey()     // user enters the API Key...
	apodQuery.DateString = getTodaysDate() // Default: Today's date!
	apodQuery.HD = "True"                  // Default: True (if applicable).

	// Template preparation.
	apodTmpl = template.New("apod")                              // create a new template.
	apodTmpl, err = apodTmpl.ParseFiles("server/html/apod.html") // parse the body file.
	checkError("Error parsing file", err)                        // check for error.

	qCounter = 0 // Query counter. Tracks the number of queries.
}

// WithDate func will become a handler for /apod/YYYY-MM-DD dates.
func WithDate(w http.ResponseWriter, r *http.Request) {
	var (
		rd       *RespData      // to hold the JSON data returned, in a Go struct..
		path     string         // the entire URL path.
		param    string         // the query parameter, passed by the user.
		idx      int            // indexing the parameter.
		resp     *http.Response // create an *http.Response var.
		respBody []byte         // the response body
		err      error          // handling the errors.
	)
	/* We need to handle the followiing: /apod/:date.
	   Slice [6:] will get the string beginning right after the "/apod/" */
	param = r.URL.Path[6:]          // get everything right after the last "/".
	idx = strings.Index(param, "/") // the index of first instance of "/" in parameter

	/* If there's nothing after "/apod/", that means NO date is entered
	   and so, we're redirecting to todays "/apod" route. */
	if len(param) == 0 {
		http.Redirect(w, r, "/apod", http.StatusMovedPermanently)
		return
	}

	/* if idx IS NOT -1, then "/" is present in parameter after the :date
	   (e.g. /apod/1968-12-28/ ) we need to redirect and remove the trailing slash. */
	if idx != -1 {
		http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusMovedPermanently)
		return
	}

	/* if idx IS -1, then there's no "/" present in parameter and we're ok. */
	if idx == -1 {
		apodQuery.DateString = param                    // date param written in query.
		path = r.URL.Path                               // the server route path.
		log.Printf("Route Path: %v\n", path)            // print route path.
		resp = makeRequest(apodQuery)                   // make request and get response.
		respBody = processTheResponse(resp)             // Process the response.
		err = json.Unmarshal(respBody, &rd)             // Unmarshal to the struct.
		checkError("Error unmarshaling JSON data", err) // check for error.
		err = apodTmpl.ExecuteTemplate(w, "apod", rd)   // execute: write to w.
		checkError("Error executing template", err)     // check for template errors.
	}
}

// makeRequest is performing a request  and select the response.
func makeRequest(qr Query) *http.Response {
	var (
		client     *http.Client   // the client which will be used for access.
		req        *http.Request  // for constructing the request.
		resp       *http.Response // to get the response.
		qrTxt      string         // the string version of the Query.
		err        error          // handling the errors.
		qv         url.Values     // the values when we run Query.
		extAPIPath string         // the external API path.
	)
	/* Create a Request: */
	// Call the method to build the Query string:
	qrTxt = qr.BuildQuery()
	req, err = http.NewRequest("GET", URL+qrTxt, nil) // create Get request.
	checkError("Error in New Request: ", err)         // error checking.
	client = &http.Client{}                           // init a client object.

	/* Send a Request and collect the response. */
	resp, err = client.Do(req)                    // make the request
	checkError("Response returned an error", err) // error checking.

	qCounter++                                // increase the counter.
	log.Printf("Run Query #%v -->", qCounter) // print the query number.

	/* Build the external API URL path. */
	// We're EXPLICITELY building the External API Path here, because if we take it from
	// the qv ParseQuery map below, it contains also the API Key.
	extAPIPath = req.URL.Scheme + "://" + req.URL.Host + req.URL.Path

	/* Parse the query Values. (map[string][]string) add the ext. API path and return. */
	qv, err = url.ParseQuery(req.URL.String())
	checkError("Error parsing query", err)
	log.Printf("URL Query:\nDate: %s HD: %s External URL API Path: %v\n",
		qv["date"], qv["hd"], extAPIPath)

	return resp // return the response.
}

// ProcessTheResponse processes server response and returns Unmarshalled data to struct.
func processTheResponse(resp *http.Response) []byte {
	var (
		rateLm         int         // rateLm the API rate limit set by NASA.
		usageDet       int         // usageDet The current remaining API calls.
		respBody       []byte      // the response body.
		respHeader     http.Header // a map[string][]string header is returned here.
		respStatus     string      // the response status.
		jsonDataFields []string    // jsonDataFields holds the JSON response fields.
		err            error       // handling the errors.
	)
	/* Get Body and Header */
	respBody, err = ioutil.ReadAll(resp.Body)      // read the response body.
	checkError("Error reading response body", err) // check for response body errors.
	defer resp.Body.Close()                        // defer close the body
	respHeader = resp.Header                       // read the entire header.

	/* Check the rate and usage limits and log them. */
	log.Printf("%s API rate & usage limits:\n", "APOD")
	rateLm, _ = strconv.Atoi(respHeader.Get("X-Ratelimit-Limit"))       // rate.
	usageDet, _ = strconv.Atoi(respHeader.Get("X-Ratelimit-Remaining")) // usage.
	log.Printf("API Hourly Rate Limit: %d\n", rateLm)
	log.Printf("API Usage Details: %d calls made, %d calls or %.2f%% of total calls remaining.\n", (rateLm - usageDet), usageDet, ((float64(usageDet) / float64(rateLm)) * 100))

	/* Check the response status */
	// NEEDS MORE WORK HERE FOR 4xx, 5xx RESPONSES.
	respStatus = resp.Status // get the response status
	log.Printf("%s API Response Status: %v\n\n", "APOD", respStatus)

	/* Find how many fields the JSON response has. */
	jsonDataFields = jsonKeys(respBody)
	log.Printf("The #fields in JSON respond is: %d\n-----------------\n\n", len(jsonDataFields))
	return respBody
}

// JSONKeys is extracting the JSON keys by creating a map, and count them.
func jsonKeys(rBody []byte) []string {
	var (
		jsonToMap map[string]string // the map used to UnMarshal the JSON.
		respKeys  []string          // holds the keys of the map (the JSON field names).
		err       error             // handling the errors.
		jsonKeys  string            // for range over the json keys.
	)
	/* Unmarshalling the body to map */
	err = json.Unmarshal(rBody, &jsonToMap)
	checkError("Error unmarshaling response body", err) // check for JSON unmarshaling.

	/* Extract map keys from JSON. */
	for jsonKeys = range jsonToMap {
		respKeys = append(respKeys, jsonKeys)
	}

	sort.Strings(respKeys) // sort keys
	log.Printf("Response JSON Fields (keys):\n%s\n", respKeys)

	return respKeys // return the keys.
}

// GetApodAPIKey prompts the API KEY from user OR assigns DEMO_KEY.
func GetApodAPIKey() string {
	var (
		n      int           // #of chars entered when scanning the API Key.
		auth   bool          // authenticate the key of the API.
		key    string        // this will hold the API Key to be returned.
		err    error         // handling the errors.
		reader *bufio.Reader // to read the key from Stdin.
	)

	/* Scan API Key: */
	log.Print("Please enter your NASA API key and press enter/return.\nIf you're not having a key, just press enter/return and \"DEMO_KEY\"\nwith limited access will be used: ")

	reader = bufio.NewReader(os.Stdin)  // create a new reader.
	key, err = reader.ReadString('\n')  // read from Stdin until enter is pressed.
	key = strings.TrimSuffix(key, "\n") // trim the last enter character ("\n")
	n = len(key)                        // get the length of the key.
	fmt.Println("# of Chars: ", n)      // print the number of chars.
	checkError("An error occured while reading input. Please try again", err)

	/* Check for null key, authentication and return: */
	// if no key is entered...
	if n == 0 {
		key = "DEMO_KEY" // ...use "DEMO_KEY" as API Key.
	}

	// If not null, check for authentication and if that fails also, pass "DEMO_KEY".
	if auth = checkApodAPIKey(URL + "?api_key=" + key); auth == false {
		log.Println("Error in key value. DEMO_KEY will be used instead.")
		key = "DEMO_KEY" // ...use "DEMO_KEY" as API Key.
	}

	// Print the value and return.
	fmt.Printf("The key to be used is: %v\n", key)
	return key
}

// CheckApodAPIKey tests the user key to the APOD API.
func checkApodAPIKey(testCall string) bool {
	var respStatusCode int // The response status code.

	/* Perform a test call to authenticate the key. */
	// If 200, key is ok. For any other kind of error, reutrn false.
	if respStatusCode = makeTestCall(testCall); respStatusCode != 200 {
		log.Printf("Response Status Code: %v\n", respStatusCode)
		return false
	}
	return true // If authentication passed, return true.
}

// MakeTestCall performs a test request to the API and returns the server's StatusCode.
func makeTestCall(test string) int {
	var (
		client *http.Client   // a client object.
		resp   *http.Response // the server response.
		err    error          // handling the errors.
	)
	/* Perform a test call to authenticate the key. */
	client = &http.Client{}         // init the client object.
	resp, err = client.Get(test)    // make a request.
	checkError("Error in key", err) // check for errors.
	return resp.StatusCode          // return the response code.
}

func getTodaysDate() string {
	var (
		timeToday    string // will hold the full time.
		dateAsString string // will hold only the date.
	)

	// Gets the full time of today as string
	timeToday = time.Now().String() // "2006-01-02 15:04:05.999999999 -0700 MST"
	dateAsString = timeToday[:10]   // the date only as string: "2006-01-02"
	return dateAsString             // return the date
}

// CheckError is for checking errors.
func checkError(errMess string, err error) {
	if err != nil {
		log.Fatalf("%s: %s\n", errMess, err)
	}
}
