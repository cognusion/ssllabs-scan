/*
 * Licensed to Qualys, Inc. (QUALYS) under one or more
 * contributor license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * QUALYS licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
*/

package main

import "crypto/tls"
import "encoding/json"
import "encoding/csv"
import "flag"
import "fmt"
import "io/ioutil"
import "log"
import "math/rand"
import "net/http"
import "strconv"
import "strings"
import "sync/atomic"
import "time"

const (
	LOG_NONE     = -1
	LOG_EMERG    = 0
	LOG_ALERT    = 1
	LOG_CRITICAL = 2
	LOG_ERROR    = 3
	LOG_WARNING  = 4
	LOG_NOTICE   = 5
	LOG_INFO     = 6
	LOG_DEBUG    = 7
	LOG_TRACE    = 8
)

var USER_AGENT = "ssllabs-scan v0.1"

var logLevel = LOG_INFO

var activeAssessments = 0

var maxAssessments = 1

var requestCounter uint64 = 0

var apiLocation = ""

var httpClient *http.Client

type LabsError struct {
	Field   string
	Message string
}

type LabsErrorResponse struct {
	ResponseErrors []LabsError `json:"errors"`
}

func (e LabsErrorResponse) Error() string {
	msg, err := json.Marshal(e)
	if (err != nil) {
		return err.Error()
	} else {
		return string(msg)
	}
}

type LabsKey struct {
	Size       int
	Strength   int
	Alg        string
	DebianFlaw bool
	Q          int
}

type LabsCert struct {
	Subject          string
	CommonNames      []string
	AltNames         []string
	NotBefore        int
	NotAfter         int
	IssuerSubject    string
	SigAlg           string
	IssuerLabel      string
	RevocationInfo   int
	CrlURIs          []string
	OcspURIs         []string
	RevocationStatus int
	Sgc              bool
	ValidationType   string
	Issues           int
}

type LabsChainCert struct {
	Subject       string
	Label         string
	IssuerSubject string
	IssuerLabel   string
	Issues        int
	Raw           string
}

type LabsChain struct {
	Certs  []LabsChainCert
	Issues int
}

type LabsProtocol struct {
	Id               int
	Name             string
	Version          string
	V2SuitesDisabled bool
	ErrorMessage     bool
	Q                int
}

type LabsSimClient struct {
	Id       int
	Name     string
	Platform string
	Version  string
	IsModern bool
}

type LabsSimulation struct {
	Client     LabsSimClient
	ErrorCode  int
	Attempts   int
	ProtocolId int
	SuiteId    int
}

type LabsSimDetails struct {
	Results []LabsSimulation
}

type LabsSuite struct {
	Id             int
	Name           string
	CipherStrength int
	DhStrength     int
	DhP            int
	DhG            int
	DhYs           int
	EcdhBits       int
	EcdhStrength   int
	Q              int
}

type LabsSuites struct {
	List       []LabsSuite
	Preference bool
}

type LabsEndpointDetails struct {
	HostStartTime      int
	Key                LabsKey
	Cert               LabsCert
	Chain              LabsChain
	Protocols          LabsProtocol
	Suites             LabsSuites
	ServerSignature    string
	PrefixDelegation   bool
	NonPrefixDelegtion bool
	VulnBeast          bool
	RenegSupport       int
	StsResponseHeader  string
	StsMaxAge          int
	StsSubdomains      bool
	PkpResponseHeader  string
	SessionResumption  int
	CompressionMethods int
	SupportsNpn        bool
	NpnProtocols       string
	SessionTickets     int
	OcspStapling       bool
	SniRequired        bool
	HttpStatusCode     int
	HttpForwarding     string
	SupportsRc4        bool
	ForwardSecrecy     int
	Rc4WithModern      bool
	Sims               LabsSimDetails
	Heartbleed         bool
	Heartbeat          bool
	OpenSslCcs         int
}

type LabsEndpoint struct {
	IpAddress            string
	ServerName           string
	StatusMessage        string
	StatusDetailsMessage string
	Grade                string
	HasWarnings          bool
	IsExceptional        bool
	Progress             int
	Duration             int
	Eta                  int
	Delegation           int
	Details              LabsEndpointDetails
}

type LabsReport struct {
	Host            string
	Port            int
	Protocol        string
	IsPublic        bool
	Status          string
	StatusMessage   string
	StartTime       int
	TestTime        int
	EngineVersion   string
	CriteriaVersion string
	CacheExpiryTime int
	Endpoints       []LabsEndpoint
	CertHostnames   []string
}

type LabsResults struct {
	reports []LabsReport
}

type LabsInfo struct {
	EngineVersion        string
	CriteriaVersion      string
	ClientMaxAssessments int
}

type SSLPulse struct {
	domainName				string
    domainDepth				int
    ipAddress				string
    port					int
    checkTime				int
    handshakeDuration		int
	validationDuration		int
    requestDuration			int
    configCheckDuration		int
    subject					string
    subjectCommonName		string
    subjectCountry			string
    altNames				string
    altNameCount			int
    isWildcard				bool
    prefixSupport			bool
    issuer					string
    issuerCommonName		string
    notAfter				int
    notBefore				int
    signatureAlg			string
    keyAlg					string
    keySize					int
    validationType			string
    isScg					int
    revocationInfo			int
    chainLength				int
    chainSize 				int
	chainIssuers			string
    chainData				string
    isTrusted				string
    supports_SSL2_hello		int
    supports_SSL_2_0		bool
    supports_SSL_3_0		bool
    supports_TLS_1_0		bool
    supports_TLS_1_1		bool
    supports_TLS_1_2		bool
    suites					string
    suiteCount				int
    suitesInOrder			bool
    supports_no_bits		bool
    supports_low_bits		bool
    supports_128_bits		bool
    supports_256_bits		bool
    em_ssl2					bool
    em_40_bits				bool
    em_56_bits				bool
    em_64_bits				bool          
    serverSignature			string
    grade					int
    gradeLetter				string
    debianFlawed			string
    sessionResumption		int
    stsResponseHeader		string
    renegSupport			int
    toleranceMinorLow		int
    toleranceMinorHigh		int
    toleranceMajorHigh		int
    pciReady				bool
    fipsReady				bool
    chainIssues				int
    fixedChainLength		int
    fixedChainSize			int
    trustAnchor				string
    engineVersion			string
    criteriaVersion			string
    crlUris					string
    ocspUris				string
    vulnBEAST				bool
    stsMaxAge				int
    stsIncludeSubdomains	bool
    pkpResponseHeader		string
    surveyId				string
    compression 			int
    npnSupport				bool
    npnProtocols			string
    sessionTickets			int
    ocspStapling			bool
    sniRequired				bool
    httpStatusCode			int
    httpForwarding			string
    keyStrength				int
    chainCount				int
    chainData2				string
    chainData3				string
    rg2009b_grade			int
    rg2009b_letter			string
    beastSuites				string
    protocolIntolerance		int
    miscIntolerance			int
    sims					string
    forwardSecrecy			int
    rc4						int
    hasWarnings				bool
    isExceptional			bool
    heartbeat				bool
    heartbleed				bool
    cve_2014_0224			int
}

func invokeGetRepeatedly(url string) (*http.Response, []byte, error) {
	retryCount := 0

	for {
		var reqId = atomic.AddUint64(&requestCounter, 1)

		if logLevel >= LOG_DEBUG {
			log.Printf("[DEBUG] Request (%v): %v", reqId, url)
		}

		req, err := http.NewRequest("GET", url, nil)
		if (err != nil) {
			return nil, nil, err
		}

		req.Header.Add("User-Agent", USER_AGENT)

		resp, err := httpClient.Do(req)
		if (err == nil) {
			// Adjust maximum concurrent requests.
			
			headerValue := resp.Header.Get("X-ClientMaxAssessments")
			if (headerValue != "") {
				i, err := strconv.Atoi(headerValue)
				if (err == nil) {
					if (maxAssessments != i) {
						maxAssessments = i
					
						if (logLevel >= LOG_INFO) {
							log.Printf("[INFO] Server set max concurrent assessments to %v", headerValue)
						}
					}
				} else {
					if (logLevel >= LOG_WARNING) {
						log.Printf("[WARNING] Ignoring invalid X-ClientMaxAssessments value (%v): %v", headerValue, err)
					}
				}
			}
			
			// Retrieve the response body.
			
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, nil, err
			}

			if logLevel >= LOG_TRACE {
				log.Printf("[TRACE] Response (%v):\n%v", reqId, string(body))
			}

			return resp, body, nil
		} else {
			if err.Error() == "EOF" {
				// Server closed a persistent connection on us, which
				// Go doesn't seem to be handling well. So we'll try one
				// more time.
				if retryCount > 1 {
					log.Fatalf("[ERROR] Too many HTTP requests failed with EOF")
				}
			} else {
				if (retryCount > 5) {
					log.Fatalf("[ERROR] Too many failed HTTP requests")
				}
				
				time.Sleep(30 * time.Second)
			}

			retryCount++
		}
	}
}

func invokeApi(command string) (*http.Response, []byte, error) {
	var url = apiLocation + "/" + command
	
	for {
		resp, body, err := invokeGetRepeatedly(url)
		if err != nil {
			return nil, nil, err
		}
		
		// Status codes 429, 503, and 529 essentially mean try later. Thus,
		// if we encounter them, we sleep for a while and try again.
		if (resp.StatusCode == 429)||(resp.StatusCode == 503) {
			if (logLevel >= LOG_NOTICE) {
				log.Printf("[NOTICE] Sleeping for 5 minutes after a %v response", resp.StatusCode)
			}
			
			time.Sleep(5 * time.Minute)	
		} else if resp.StatusCode == 529 {
			// In case of the overloaded server, randomize the sleep time so
			// that some clients reconnect earlier and some later.
			
			sleepTime := 15 + rand.Int31n(15)
			
			if (logLevel >= LOG_NOTICE) {
				log.Printf("[NOTICE] Sleeping for %v minutes after a 529 response", sleepTime)
			}
			
			time.Sleep(time.Duration(sleepTime) * time.Minute)
		} else if (resp.StatusCode != 200) && (resp.StatusCode != 400) {
			log.Fatalf("[ERROR] Unexpected response status code %v", resp.StatusCode)
		} else {
			return resp, body, nil
		}
	}
}

func invokeInfo() (*LabsInfo, error) {
	var command = "info"

	_, body, err := invokeApi(command)
	if err != nil {
		return nil, err
	}

	var labsInfo LabsInfo
	err = json.Unmarshal(body, &labsInfo)
	if err != nil {
		return nil, err
	}

	return &labsInfo, nil
}

func invokeAnalyze(host string, clearCache bool) (*LabsReport, error) {
	var command = "analyze?host=" + host + "&all=done"

	if clearCache {
		command = command + "&clearCache=on"
	}

	resp, body, err := invokeApi(command)
	if err != nil {
		return nil, err
	}

	// Use the status code to determine if the response is an error.
	if (resp.StatusCode == 400) {
		// Parameter validation error.
		
		var apiError LabsErrorResponse
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return nil, err
		}
		
		return nil, apiError
	} else {
		// We should have a proper response.
		
		var analyzeResponse LabsReport
		err = json.Unmarshal(body, &analyzeResponse)
		if err != nil {
			return nil, err
		}

		return &analyzeResponse, nil
	}
}

type Event struct {
	host      string
	eventType int
	report    *LabsReport
}

const (
	ASSESSMENT_STARTING = 0
	ASSESSMENT_COMPLETE = 1
)

func NewAssessment(host string, eventChannel chan Event) {
	eventChannel <- Event { host, ASSESSMENT_STARTING, nil }

	var report *LabsReport
	var clearCache = true
	var startTime = -1

	for {
		myResponse, err := invokeAnalyze(host, clearCache)
		if err != nil {
			log.Fatalf("[ERROR] Assessment failed: %v", err)
		}

		if clearCache == true {
			clearCache = false
			startTime = myResponse.StartTime
		} else {
			if myResponse.StartTime != startTime {
				log.Fatalf("[ERROR] Inconsistent startTime. Expected %v, got %v.", startTime, myResponse.StartTime)
			}
		}

		if (myResponse.Status == "READY") || (myResponse.Status == "ERROR") {
			report = myResponse
			break
		}

		time.Sleep(5 * time.Second)
	}

	eventChannel <- Event { host, ASSESSMENT_COMPLETE, report }
}

type HostProvider struct {
	hostnames []string
	i         int
}

func NewHostProvider(hs []string) *HostProvider {
	hostProvider := HostProvider{hs, 0}
	return &hostProvider
}

func (hp *HostProvider) next() (string, bool) {
	if hp.i < len(hp.hostnames) {
		host := hp.hostnames[hp.i]
		hp.i = hp.i + 1
		return host, true
	} else {
		return "", false
	}
}

type Manager struct {
	hostProvider         *HostProvider
	FrontendEventChannel chan Event
	BackendEventChannel  chan Event
	results              *LabsResults
}

func NewManager(hostProvider *HostProvider) *Manager {
	manager := Manager{
		hostProvider:         hostProvider,
		FrontendEventChannel: make(chan Event),
		BackendEventChannel:  make(chan Event),
		results:              &LabsResults{reports: make([]LabsReport, 0)},
	}

	go manager.run()

	return &manager
}

func (manager *Manager) startAssessment(h string) {
	go NewAssessment(h, manager.BackendEventChannel)
	activeAssessments++
}

func (manager *Manager) run() {
	// XXX Allow self-signed certificates for now. Will be removed in the final version.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config { InsecureSkipVerify: true },
    }
	
    httpClient = &http.Client { Transport: transport }

	// Ping SSL Labs to determine how many concurrent
	// assessments we're allowed to use. Print the API version
	// information and the limits.

	labsInfo, err := invokeInfo()
	if err != nil {
		// TODO Signal error so that we return the correct exit code
		close(manager.FrontendEventChannel)
	}

	if logLevel >= LOG_INFO {
		log.Printf("[INFO] SSL Labs v%v (criteria version %v)", labsInfo.EngineVersion, labsInfo.CriteriaVersion)
	}

	moreAssessments := true

	for {
		select {
		// Handle assessment events (e.g., starting and finishing).
		case e := <-manager.BackendEventChannel:
			if e.eventType == ASSESSMENT_STARTING {
				if logLevel >= LOG_INFO {
					log.Printf("[INFO] Assessment starting: %v", e.host)
				}
			}

			if e.eventType == ASSESSMENT_COMPLETE {
				if logLevel >= LOG_INFO {
					msg := ""

					// Missing C's ternary operator here.
					if len(e.report.Endpoints) == 0 {
						msg = fmt.Sprintf("[INFO] Assessment failed: %v (%v)", e.host, e.report.StatusMessage)
					} else if len(e.report.Endpoints) > 1 {
						msg = fmt.Sprintf("[INFO] Assessment complete: %v (%v hosts in %v seconds)",
							e.host, len(e.report.Endpoints), (e.report.TestTime-e.report.StartTime)/1000)
					} else {
						msg = fmt.Sprintf("[INFO] Assessment complete: %v (%v host in %v seconds)",
							e.host, len(e.report.Endpoints), (e.report.TestTime-e.report.StartTime)/1000)
					}

					for _, endpoint := range e.report.Endpoints {
						if (endpoint.Grade != "") {
							msg = msg + "\n    " + endpoint.IpAddress + ": " + endpoint.Grade
						} else {
							msg = msg + "\n    " + endpoint.IpAddress + ": Err: " + endpoint.StatusMessage
						}
					}

					log.Println(msg)
				}

				activeAssessments--

				manager.results.reports = append(manager.results.reports, *e.report)

				// Are we done?
				if (activeAssessments == 0) && (moreAssessments == false) {
					close(manager.FrontendEventChannel)
					return
				}
			}

			break

		// Once a second, start a new assessment, provided there are
		// hostnames left and we're not over the concurrent assessment limit.
		default:
			<-time.NewTimer(time.Second).C
			if moreAssessments {
				if activeAssessments < maxAssessments {
					host, hasNext := manager.hostProvider.next()
					if hasNext {
						manager.startAssessment(host)
					} else {
						// We've run out of hostnames and now just need
						// to wait for all the assessments to complete.
						moreAssessments = false
					}
				}
			}
			break
		}
	}
}

func parseLogLevel(level string) int {
	switch {
	case level == "error":
		return LOG_ERROR
	case level == "info":
		return LOG_INFO
	case level == "debug":
		return LOG_DEBUG
	case level == "trace":
		return LOG_TRACE
	}

	log.Fatalf("[ERROR] Unrecognized log level: %v", level)
	return -1
}

func main() {
	var conf_api = flag.String("api", "REQUIRED", "API entry point, for example https://www.example.com/api/")
	var conf_verbosity = flag.String("verbosity", "info", "Configure log verbosity: error, info, debug, or trace.")
	var conf_json_pretty = flag.Bool("json-pretty", false, "Enable pretty JSON output")
	var conf_quiet = flag.Bool("quiet", false, "Disable status messages (logging)")
	var conf_csv = flag.Bool("csv", false, "Output results in SSL Pulse CSV format")

	flag.Parse()

	logLevel = parseLogLevel(strings.ToLower(*conf_verbosity))
	
	if (*conf_quiet) {
		logLevel = LOG_NONE
	}

	// TODO Verify that the API entry point is a URL.
	apiLocation = *conf_api

	// TODO Validate all hostnames before we attempt to test them. At least
	//      one hostname is required.
	hostnames := flag.Args()

	hp := NewHostProvider(hostnames)
	manager := NewManager(hp)

	// Respond to events until all the work is done.
	for {
		_, running := <-manager.FrontendEventChannel
		if running == false {
			var results []byte
			var err error

			if *conf_json_pretty {
				results, err = json.MarshalIndent(manager.results.reports, "", "    ")
			} else {
				results, err = json.Marshal(manager.results.reports)
			}

			if err != nil {
				log.Fatalf("[ERROR] Output to JSON failed: %v", err)
			}

			fmt.Println(string(results))

			if logLevel >= LOG_INFO {
				log.Println("[INFO] All assessments complete; shutting down")
			}

			return
		}
	}
}
