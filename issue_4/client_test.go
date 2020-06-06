package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
	"strings"
)

const filePath string = "./dataset.xml"

// код писать тут

type XmlRow struct{
	FirstName string  `xml:"first_name" json:"-"`
	LastName string   `xml:"last_name" json:"-"`
	Id int            `xml:"id"`
	Age int           `xml:"age"`
	About string      `xml:"about"`
	Name string 	  `xml:"-"`
	Gender string     `xml:"gender"`
}

type XmlRows struct {
    List    []XmlRow `xml:"row"`
}

type By func(row1, row2 *XmlRow) bool

func (by By) Sort(rows []XmlRow) {
	rs := &rowSorter{
		rows: rows,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(rs)
}

type rowSorter struct {
	rows []XmlRow
	by   func(row1, row2 *XmlRow) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *rowSorter) Len() int {
	return len(s.rows)
}

// Swap is part of sort.Interface.
func (s *rowSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *rowSorter) Less(i, j int) bool {
	return s.by(&s.rows[i], &s.rows[j])
}

func (rows *XmlRows) SortXmlRows(OrderField string, OrderBy int) error {
	var less func(row1, row2 *XmlRow) bool
	switch OrderField {
	case "", "Name":
		less = func(row1, row2 *XmlRow) bool {
			return row1.Name < row2.Name
		}
	case "Age":
		less = func(row1, row2 *XmlRow) bool {
			return row1.Age < row2.Age
		}
	case "Id":
		less = func(row1, row2 *XmlRow) bool {
			return row1.Id < row2.Id
		}
	default:
		return errors.New(ErrorBadOrderField)
	}

	switch OrderBy {
	case OrderByAsc:
		By(less).Sort(rows.List)
	case OrderByAsIs:
		return nil
	case OrderByDesc:
		more := func(row1, row2 *XmlRow) bool{
			return less(row2, row1)
		}
		By(more).Sort(rows.List)
	default:
		return errors.New(ErrorBadOrderField)
	}
	return nil
}

func SearchServer(w http.ResponseWriter, r *http.Request){

	if r.Header.Get("AccessToken") != "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	v := new(XmlRows)
	err = xml.Unmarshal(fileContents, &v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := range v.List{
		v.List[i].Name = v.List[i].FirstName + v.List[i].LastName
	}

	q:= r.FormValue("query")
	if q != ""{
		var List []XmlRow
		for _, r := range v.List{
			if strings.Contains(r.About, q){
				List = append(List, r)
				continue
			}
		}
		v.List = List
	}

	OrderField := r.FormValue("order_field")
	OrderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error":"ErrorBadOrderField"}`))
		return
	}
	err = v.SortXmlRows(OrderField, OrderBy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error":"ErrorBadOrderField"}`))
		return
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error":"ErrorBadOrderField"}`))
		return
	}

	if limit < len(v.List){
		v.List = v.List[0:limit]
	}

	jsonResponse, err := json.Marshal(&v.List)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func TimeOutServer(w http.ResponseWriter, r *http.Request){
	time.Sleep(client.Timeout * 2)
	w.WriteHeader(http.StatusOK)
}

func ErrorJsonOnBadRequestServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`"IT IS NOT JSON"`))
}

func UnknownBadRequestServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"Error":"Unknown"}`))
}

func ErrorJsonOnOkStatusServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`"IT IS NOT JSON"`))
}

func InternalErrorServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusInternalServerError)
}



type TestCase struct{
	SearchReq SearchRequest
	Result *SearchResponse
	IsError bool
	SearchServer func(w http.ResponseWriter, r *http.Request)
	AccessToken string
}

func TestFindXmlRows(t *testing.T){
	cases := []TestCase{
		TestCase{
			SearchReq: SearchRequest{
				OrderField: "BadField",
			},
			IsError: true,
			SearchServer:SearchServer,
		},
		TestCase{
			SearchReq: SearchRequest{
				Limit:      -1,
			},
			IsError: true,
			SearchServer:SearchServer,
		},
		TestCase{
			SearchReq: SearchRequest{
				Offset:      -1,
			},
			IsError: true,
			SearchServer:SearchServer,
		},
		TestCase{
			IsError: true,
			SearchServer:TimeOutServer,
		},
		TestCase{
			IsError: true,
			SearchServer:ErrorJsonOnBadRequestServer,
		},
		TestCase{
			IsError: true,
			SearchServer:UnknownBadRequestServer,
		},
		TestCase{
			IsError: true,
			SearchServer:ErrorJsonOnOkStatusServer,
		},
		TestCase{
			IsError: true,
			SearchServer:InternalErrorServer,
		},
		TestCase{
			AccessToken: "BadAccessToken",
			IsError: true,
			SearchServer:SearchServer,
		},
		TestCase{
			SearchReq: SearchRequest{
				OrderField: "BadField",
			},
			IsError: true,
			SearchServer:SearchServer,
		},
		TestCase{
			SearchReq:    SearchRequest{
				Limit:      1,
			},
			Result:      &SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "BoydWolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			} ,
			SearchServer: SearchServer,
		},
		TestCase{
			SearchReq:    SearchRequest{
				Limit:      50,
				Query:      "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat",
			},
			Result:      &SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "BoydWolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
			} ,
			SearchServer: SearchServer,
		},
	}

	for num, c := range cases{
		ts := httptest.NewServer(http.HandlerFunc(c.SearchServer))
		sc := &SearchClient{URL: ts.URL, AccessToken: c.AccessToken}
		res, err := sc.FindUsers(c.SearchReq)
		if err != nil && !c.IsError{
			t.Errorf("[%d] unexpected error: %#v", num, err)
		}
		if err == nil && c.IsError{
			t.Errorf("[%d] expected error, got nil", num)
		}
		if !reflect.DeepEqual(c.Result, res) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v",
				num, c.Result, res)}
		ts.Close()
	}

}


func TestBadURL(t *testing.T) {
    SearchReq:= SearchRequest{}
	sc := &SearchClient{URL: "It was very, very bad URL"}
	_, err := sc.FindUsers(SearchReq)
	if err == nil{
		t.Errorf("Expected error, got nil")
	}
}


//func TestWrongURL(t *testing.T){
//	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
//}