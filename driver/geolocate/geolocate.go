// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package geolocate implements an interface
// to GEOLocate gazetteer.
package geolocate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/geography"

	"github.com/pkg/errors"
)

// Retry is the number of times a request will be retried
// before aborted.
var Retry = 5

// Timeout is the timeout of the http request.
var Timeout = 20 * time.Second

// Wait is the waiting time for a new request
// (we don't want to overload the GEOLocate server!).
var Wait = time.Millisecond * 300

// Buffer is the maximum number of requests in the request queue.
var Buffer = 100

const wsHead = "http://www.museum.tulane.edu/webservices/geolocatesvcv2/glcwrap.aspx?"

// Request contains a GEOLocate request,
// and a channel with the answers.
type request struct {
	req string
	ans chan bytes.Buffer
	err chan error
}

// NewRequest sends a request to the request channel.
func newRequest(req string) request {
	r := request{
		req: wsHead + req,
		ans: make(chan bytes.Buffer),
		err: make(chan error),
	}
	reqChan.cReqs <- r
	return r
}

// ReqChanType keeps the requests channel.
type reqChanType struct {
	cReqs chan request
}

// ReqChan is the requests channel.
// It should be initialized before
// using the database.
var reqChan *reqChanType

// InitReqs initialize the request channel.
func initReqs() {
	http.DefaultClient.Timeout = Timeout
	reqChan = &reqChanType{cReqs: make(chan request, Buffer)}
	go reqChan.reqs()
}

// Reqs make the network request.
func (rc *reqChanType) reqs() {
	for r := range rc.cReqs {
		a, err := http.Get(r.req)
		if err != nil {
			r.err <- err
			continue
		}
		var b bytes.Buffer
		b.ReadFrom(a.Body)
		a.Body.Close()
		r.ans <- b

		// we do not want to overload the gbif server.
		time.Sleep(Wait)
	}
}

func init() {
	biodv.RegisterGz("geolocate", biodv.GzDriver{Open, aboutGEOLocate})
}

// AboutGEOLocate returns a simple statement of the purpose of the driver.
func aboutGEOLocate() string {
	return "a driver for GEOLocate gazetteer service"
}

// Open returns the GEOLocate service handle
// that implements the biodv.Gazetteer interface.
func Open(param string) (biodv.Gazetteer, error) {
	if reqChan == nil {
		initReqs()
	}
	return gzService{}, nil
}

// GzService is a biodv.Gazetteer.
type gzService struct{}

func (gz gzService) Locate(adm geography.Admin, locality string) *biodv.GeoScan {
	sc := biodv.NewGeoScan(100)
	adm.Country = strings.ToUpper(adm.Country)
	if !geography.IsValidCode(adm.Country) {
		sc.Add(geography.NewPosition(), errors.Errorf("geolocate: A valid country must be given to Locate"))
		return sc
	}
	nm, ok := country[adm.Country]
	if !ok {
		sc.Add(geography.NewPosition(), errors.Errorf("geolocate: country %s not on GEOLocate", geography.Country(adm.Country)))
		return sc
	}
	if nm == "UNITED STATES OF AMERICA" {
		switch adm.Country {
		case "VI":
			adm.State = "Virgin Islands"
		case "AS":
			adm.State = "American Samoa"
		case "PR":
			adm.State = "Puerto Rico"
		case "GU":
			adm.State = "Guam"
		}
		if adm.State == "" {
			sc.Add(geography.NewPosition(), errors.Errorf("geolocate: in USA, GEOLocate only search when state is included"))
			return sc
		}
	}
	locality = strings.TrimSpace(locality)
	if locality == "" {
		sc.Add(geography.NewPosition(), errors.Errorf("geolocate: Empty locality"))
		return sc
	}

	param := url.Values{}
	param.Add("coutry", nm)
	param.Add("locality", locality)
	if adm.State != "" {
		param.Add("state", adm.State)
	}
	if adm.County != "" {
		param.Add("county", adm.County)
	}
	param.Add("enableH20", "false")
	param.Add("hwyX", "false")
	param.Add("fmt", "geojson")
	go gz.pointList(sc, param)
	return sc
}

func (gz gzService) Reverse(p geography.Position) (geography.Admin, error) {
	return geography.Admin{}, nil
}

type locAnswer struct {
	Features []feature
}

type feature struct {
	Geometry   point
	Properties property
}

type point struct {
	Coordinates []float64
}

type property struct {
	UncertaintyRadiusMeters uint
	Debug                   string
}

// PointList returns an specific list of points.
func (gz gzService) pointList(sc *biodv.GeoScan, param url.Values) {
	var err error
	for r := 0; r < Retry; r++ {
		req := newRequest(wsHead + param.Encode())
		select {
		case err = <-req.err:
			continue
		case a := <-req.ans:
			var resp *locAnswer
			resp, err = decodePointList(&a)
			if err != nil {
				continue
			}

			for _, f := range resp.Features {
				p := geography.Position{
					Lat:         f.Geometry.Coordinates[1],
					Lon:         f.Geometry.Coordinates[0],
					Uncertainty: f.Properties.UncertaintyRadiusMeters,
					Source:      "web:geolocate",
				}
				if !sc.Add(p, nil) {
					return
				}
			}
			r = Retry
		}
	}
	if err != nil {
		sc.Add(geography.NewPosition(), errors.Wrapf(err, "geolocate"))
		return
	}
	sc.Add(geography.NewPosition(), nil)
}

func decodePointList(b *bytes.Buffer) (*locAnswer, error) {
	d := json.NewDecoder(b)
	resp := &locAnswer{}
	err := d.Decode(resp)
	return resp, err
}

// Country is a valid ISO 3166-1 alpha-2 county code,
// and its common name.
//
// Countries commented,
// are not available in geolocate.
var country = map[string]string{
	// "AD": "Andorra",
	"AE": "UNITED ARAB EMIRATES",
	"AF": "AFGHANISTAN",
	// "AG": "Antigua and Barbuda",
	"AI": "ANGUILLA",
	"AL": "ALBANIA",
	"AM": "ARMENIA",
	"AO": "ANGOLA",
	// "AQ": "Antarctica",
	"AR": "ARGENTINA",
	"AS": "UNITED STATES OF AMERICA",
	"AT": "AUSTRIA",
	"AU": "AUSTRALIA",
	"AW": "ARUBA",
	// "AX": "Aland Islands",
	"AZ": "AZERBAIJAN",
	"BA": "BOSNIA AND HERZEGOVINA",
	"BB": "BARBADOS",
	"BD": "BANGLADESH",
	"BE": "BELGIUM",
	"BF": "BURKINA FASO",
	"BG": "BULGARIA",
	"BH": "BAHRAIN",
	"BI": "BURUNDI",
	"BJ": "BENIN",
	// "BL": "Saint Barthelemy",
	"BM": "BERMUDA",
	"BN": "BRUNEI",
	"BO": "BOLIVIA",
	"BQ": "NETHERLANDS ANTILLES",
	"BR": "BRAZIL",
	"BS": "BAHAMAS",
	"BT": "BHUTAN",
	"BV": "BOUVET ISLAND",
	"BW": "BOTSWANA",
	"BY": "BELARUS",
	"BZ": "BELIZE",
	"CA": "CANADA",
	"CC": "COCOS (KEELING) ISLANDS",
	"CD": "CONGO, DEMOCRATIC REPUBLIC OF THE",
	"CF": "CENTRAL AFRICAN REPUBLIC",
	"CG": "CONGO",
	"CH": "SWITZERLAND",
	"CI": "COTE D'IVOIRE",
	"CK": "COOK ISLANDS",
	"CL": "CHILE",
	"CM": "CAMEROON",
	"CN": "CHINA",
	"CO": "COLOMBIA",
	"CR": "COSTA RICA",
	"CU": "CUBA",
	"CV": "CAPE VERDE",
	"CW": "NETHERLANDS ANTILLES",
	"CX": "CHRISTMAS ISLAND",
	"CY": "CYPRUS",
	"CZ": "CZECH REPUBLIC",
	"DE": "GERMANY",
	"DJ": "DJIBOUTI",
	"DK": "DENMARK",
	"DM": "DOMINICA",
	"DO": "DOMINICAN REPUBLIC",
	"DZ": "ALGERIA",
	"EC": "ECUADOR",
	"EE": "ESTONIA",
	"EG": "EGYPT",
	"EH": "WESTERN SAHARA",
	"ER": "ERITREA",
	"ES": "SPAIN",
	"ET": "ETHIOPIA",
	"FI": "FINLAND",
	"FJ": "FIJI",
	"FK": "FALKLAND ISLANDS",
	"FM": "MICRONESIA, FEDERATED STATES OF",
	"FO": "FAROE ISLANDS",
	"FR": "FRANCE",
	"GA": "GABON",
	"GB": "UNITED KINGDOM",
	"GD": "GRENADA",
	"GE": "GEORGIA",
	"GF": "FRENCH GUIANA",
	"GG": "GUERNSEY",
	"GH": "GHANA",
	"GI": "GIBRALTAR",
	"GL": "GREENLAND",
	"GM": "GAMBIA, THE",
	"GN": "GUINEA",
	"GP": "GUADELOUPE",
	"GQ": "EQUATORIAL GUINEA",
	"GR": "GREECE",
	"GS": "SOUTH GEORGIA AND THE SOUTH SANDWICH ISLANDS",
	"GT": "GUATEMALA",
	"GU": "UNITED STATES OF AMERICA",
	"GW": "GUINEA-BISSAU",
	"GY": "GUYANA",
	"HK": "Hong Kong",
	"HM": "HEARD ISLAND AND MCDONALD ISLANDS",
	"HN": "HONDURAS",
	"HR": "CROATIA",
	"HT": "HAITI",
	"HU": "HUNGARY",
	"ID": "INDONESIA",
	"IE": "IRELAND",
	"IL": "ISRAEL",
	"IM": "ISLE OF MAN",
	"IN": "INDIA",
	"IO": "BRITISH INDIAN OCEAN TERRITORY",
	"IQ": "IRAQ",
	"IR": "IRAN",
	"IS": "ICELAND",
	"IT": "ITALY",
	"JE": "JERSEY",
	"JM": "JAMAICA",
	"JO": "JORDAN",
	"JP": "JAPAN",
	"KE": "KENYA",
	"KG": "KYRGYZSTAN",
	"KH": "CAMBODIA",
	"KI": "KIRIBATI",
	"KM": "COMOROS",
	"KN": "SAINT KITTS AND NEVIS",
	"KP": "NORTH KOREA",
	"KR": "SOUTH KOREA",
	"KW": "KUWAIT",
	"KY": "CAYMAN ISLANDS",
	"KZ": "KAZAKHSTAN",
	"LA": "LAOS",
	"LB": "LEBANON",
	"LC": "SAINT LUCIA",
	"LI": "LIECHTENSTEIN",
	"LK": "SRI LANKA",
	"LR": "LIBERIA",
	"LS": "LESOTHO",
	"LT": "LITHUANIA",
	"LU": "LUXEMBOURG",
	"LV": "LATVIA",
	"LY": "LIBYA",
	"MA": "MOROCCO",
	"MC": "MONACO",
	"MD": "MOLDOVA",
	"ME": "SERBIA AND MONTENEGRO",
	// "MF": "Saint Martin (French part)",
	"MG": "MADAGASCAR",
	"MH": "MARSHALL ISLANDS",
	"MK": "MACEDONIA, THE FORMER YUGOSLAV REPUBLIC OF",
	"ML": "MALI",
	"MM": "BURMA",
	"MN": "MONGOLIA",
	// "MO": "Macao",
	// "MP": "Northern Mariana Islands",
	"MQ": "MARTINIQUE",
	"MR": "MAURITANIA",
	"MS": "MONTSERRAT",
	"MT": "MALTA",
	"MU": "MAURITIUS",
	"MV": "MALDIVES",
	"MW": "MALAWI",
	"MX": "MEXICO",
	"MY": "MALAYSIA",
	"MZ": "MOZAMBIQUE",
	"NA": "NAMIBIA",
	"NC": "NEW CALEDONIA",
	"NE": "NIGER",
	"NF": "NORFOLK ISLAND",
	"NG": "NIGERIA",
	"NI": "NICARAGUA",
	"NL": "NETHERLANDS",
	"NO": "NORWAY",
	"NP": "NEPAL",
	"NR": "NAURU",
	"NU": "NIUE",
	"NZ": "NEW ZEALAND",
	"OM": "OMAN",
	"PA": "PANAMA",
	"PE": "PERU",
	"PF": "FRENCH POLYNESIA",
	"PG": "PAPUA NEW GUINEA",
	"PH": "PHILIPPINES",
	"PK": "PAKISTAN",
	"PL": "POLAND",
	"PM": "SAINT PIERRE AND MIQUELON",
	"PN": "PITCAIRN ISLANDS",
	"PR": "UNITED STATES OF AMERICA",
	"PS": "WEST BANK",
	"PT": "PORTUGAL",
	"PW": "PALAU",
	"PY": "PARAGUAY",
	"QA": "QATAR",
	"RE": "REUNION",
	"RO": "ROMANIA",
	"RS": "SERBIA AND MONTENEGRO",
	"RU": "RUSSIA",
	"RW": "RWANDA",
	"SA": "SAUDI ARABIA",
	"SB": "SOLOMON ISLANDS",
	"SC": "SEYCHELLES",
	"SD": "SUDAN",
	"SE": "SWEDEN",
	"SG": "SINGAPORE",
	"SH": "SAINT HELENA",
	"SI": "SLOVENIA",
	"SJ": "SVALBARD",
	"SK": "SLOVAKIA",
	"SL": "SIERRA LEONE",
	"SM": "SAN MARINO",
	"SN": "SENEGAL",
	"SO": "SOMALIA",
	"SR": "SURINAME",
	"SS": "SOUTH SUDAN",
	"ST": "SAO TOME AND PRINCIPE",
	"SV": "EL SALVADOR",
	"SX": "NETHERLANDS ANTILLES",
	"SY": "SYRIA",
	"SZ": "SWAZILAND",
	"TC": "TURKS AND CAICOS ISLANDS",
	"TD": "Chad",
	"TF": "FRENCH SOUTHERN AND ANTARCTIC LANDS",
	"TG": "TOGO",
	"TH": "THAILAND",
	"TJ": "TAJIKISTAN",
	"TK": "TOKELAU",
	"TL": "EAST TIMOR",
	"TM": "TURKMENISTAN",
	"TN": "TUNISIA",
	"TO": "TONGA",
	"TR": "TURKEY",
	"TT": "TRINIDAD AND TOBAGO",
	"TV": "TUVALU",
	"TW": "TAIWAN",
	"TZ": "TANZANIA",
	"UA": "UKRAINE",
	"UG": "UGANDA",
	// "UM": "United States Minor Outlying Islands",
	"US": "UNITED STATES OF AMERICA",
	"UY": "URUGUAY",
	"UZ": "UZBEKISTAN",
	"VA": "VATICAN CITY",
	"VC": "SAINT VINCENT AND THE GRENADINES",
	"VE": "VENEZUELA",
	"VG": "BRITISH VIRGIN ISLANDS",
	"VI": "UNITED STATES OF AMERICA",
	"VN": "VIETNAM",
	"VU": "VANUATU",
	"WF": "WALLIS AND FUTUNA",
	"WS": "SAMOA",
	"YE": "YEMEN",
	"YT": "MAYOTTE",
	"ZA": "SOUTH AFRICA",
	"ZM": "ZAMBIA",
	"ZW": "ZIMBABWE",
}

// The following territories are on geolocate
// but not available on ISO 3166-1
//	GLORIOSO ISLANDS
//	GAZA STRIP
//	JAN MAYEN	[Svalbard are taken]
//	JUAN DE NOVA ISLAND
//	PARACEL ISLANDS
//	SPRATLY ISLANDS
