package utils

import "testing"

func TestEncodeUserId(t *testing.T) {
	actual := EncodeUserId("U4960c75d29849705bba861ff06c70f2f")
	expected := "SWDHXSmElwW7qGH_BscPLw"
	if actual != expected {
		t.Errorf("eq: want %t got %t", expected, actual)
	}
}

func TestDecodeUserId(t *testing.T) {
	actual := DecodeUserId("SWDHXSmElwW7qGH_BscPLw")
	expected := "U4960c75d29849705bba861ff06c70f2f"
	if actual != expected {
		t.Errorf("eq: want %t got %t", expected, actual)
	}
}

func TestExtractEncodeUserId(t *testing.T) {
	actual := ExtractEncodeUserId(`Approve subscriber "Nguan ffr" (SWDHXSmElwW7qGH_BscPLw)`)
	expected := "U4960c75d29849705bba861ff06c70f2f"
	if actual != expected {
		t.Errorf("eq: want %t got %t", expected, actual)
	}
}