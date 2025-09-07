package resolver

import (
	"testing"
)

type TestData struct {
	Record      string
	Want        RR
	ErrorString string
}

func doParseRecordTest(t *testing.T, test TestData) {
	got, err := parseRecord(test.Record)

	if err != nil {
		if test.ErrorString != "" {
			if err.Error() != test.ErrorString {
				t.Errorf("parseRecord(%s) = %v, want error %v", test.Record, err, test.ErrorString)
			}
		} else {
			t.Errorf("parseRecord(%s) = %v", test.Record, err)
		}
	}

	if got != test.Want {
		t.Errorf("parseRecord(%s) = %v, want %v", test.Record, got, test.Want)
	}
}

func TestParseRecord_Success(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_Route_Preserve_Success(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; route=preserve",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, PreserveRoute: true},
	})
}

func TestParseRecord_Route_Preserve_Invalid(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; route;",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, PreserveRoute: false},
	})

	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; route=drop",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, PreserveRoute: false},
	})
}

func TestParseRecord_Code_Success(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; code=301",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, Code: 301},
	})
}

func TestParseRecord_Code_Invalid(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; code=abc",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, Code: 302},
	})

	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; code=111",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, Code: 302},
	})

	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; code",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false, Code: 302},
	})
}

func TestParseRecord_Splitting(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1;;;;   ;;; ; ; ;;; dest=https://example.com",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_QuoteTriming(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "\"v=srd1; dest=https://example.com\"",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_ExtraFields(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; extra=field; extra2=field2",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_InvalidDestUrl(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record:      "v=srd1; dest=xyz://example.^.com",
		Want:        RR{NotFound: true},
		ErrorString: "invalid destination",
	})
}

func TestParseRecord_NotFound(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record:      "v=srd1;",
		Want:        RR{NotFound: true},
		ErrorString: "no destination found",
	})
}
