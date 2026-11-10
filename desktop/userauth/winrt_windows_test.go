package userauth

import (
	"strings"
	"testing"
)

// Known IIDs published in the Windows SDK headers check the algorithm.
func TestPinterfaceIID(t *testing.T) {
	for _, c := range []struct{ sig, want string }{
		{"pinterface({9fc2b0bb-e446-44e2-aa61-9cab8f636af2};b1)", "{CDB5EFB3-5788-509D-9BE1-71CCB8A3362A}"},     // IAsyncOperation<Boolean>
		{"pinterface({9fc2b0bb-e446-44e2-aa61-9cab8f636af2};string)", "{3E1FE603-F897-5263-B328-0806426B8A79}"}, // IAsyncOperation<HSTRING>
	} {
		if got := pinterfaceIID(c.sig).String(); !strings.EqualFold(got, c.want) {
			t.Errorf("%s: got %s, want %s", c.sig, got, c.want)
		}
	}
	if got := iidAsyncOpVerificationResult.String(); !strings.EqualFold(got, "{FD596FFD-2318-558F-9DBE-D21DF43764A5}") {
		t.Errorf("IAsyncOperation<UserConsentVerificationResult> = %s", got)
	}
}
