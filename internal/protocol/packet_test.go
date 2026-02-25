package protocol

import "testing"

func TestEncodeDecodeAudioPacket(t *testing.T) {
	inSeq := uint32(42)
	inSamples := []int16{-32768, -1, 0, 1, 32767}

	packet := EncodeAudioPacket(inSeq, inSamples)
	outSeq, outSamples, err := DecodeAudioPacket(packet)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if outSeq != inSeq {
		t.Fatalf("unexpected sequence: got %d want %d", outSeq, inSeq)
	}

	if len(outSamples) != len(inSamples) {
		t.Fatalf("unexpected samples size: got %d want %d", len(outSamples), len(inSamples))
	}

	for i := range inSamples {
		if outSamples[i] != inSamples[i] {
			t.Fatalf("sample[%d]: got %d want %d", i, outSamples[i], inSamples[i])
		}
	}
}

func TestDecodeAudioPacketWithInvalidPayload(t *testing.T) {
	_, _, err := DecodeAudioPacket([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for short packet")
	}
}

func TestSequenceAhead(t *testing.T) {
	if !SequenceAhead(10, 9) {
		t.Fatal("expected 10 to be ahead of 9")
	}
	if SequenceAhead(9, 10) {
		t.Fatal("did not expect 9 to be ahead of 10")
	}
	if !SequenceAhead(1, ^uint32(0)) {
		t.Fatal("expected wrap-around sequence to be ahead")
	}
}
