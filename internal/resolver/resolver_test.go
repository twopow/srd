package resolver

import "testing"

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

func TestParseRecord_Splitting(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1;;;;   ;;; ; ; ;;; dest=https://example.com",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_ExtraFields(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record: "v=srd1; dest=https://example.com; extra=field; extra2=field2",
		Want:   RR{Version: "srd1", To: "https://example.com", NotFound: false},
	})
}

func TestParseRecord_NotFound(t *testing.T) {
	doParseRecordTest(t, TestData{
		Record:      "v=srd1;",
		Want:        RR{NotFound: true},
		ErrorString: "no destination found",
	})
}
