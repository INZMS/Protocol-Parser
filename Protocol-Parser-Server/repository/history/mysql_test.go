package history

import "testing"

func TestHashPacketCanonicalizesCaseAndOuterWhitespace(t *testing.T) {
	first := hashPacket("2929", "292980AABB")
	second := hashPacket(" 2929 ", " 292980aabb ")
	if first != second {
		t.Fatalf("equivalent packets must have the same hash: %s != %s", first, second)
	}
	if first == hashPacket("2929", "292980AABC") {
		t.Fatal("different packets must not have the same hash")
	}
	if first == hashPacket("JT808", "292980AABB") {
		t.Fatal("the protocol must be part of the packet identity")
	}
}
