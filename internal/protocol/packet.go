package protocol

import (
	"encoding/binary"
	"errors"
)

const HeaderSize = 4

var ErrPacketTooSmall = errors.New("audio packet too small")

func EncodeAudioPacket(sequence uint32, pcm []int16) []byte {
	packet := make([]byte, HeaderSize+len(pcm)*2)
	binary.LittleEndian.PutUint32(packet[:HeaderSize], sequence)

	for i, sample := range pcm {
		offset := HeaderSize + i*2
		binary.LittleEndian.PutUint16(packet[offset:offset+2], uint16(sample))
	}

	return packet
}

func DecodeAudioPacket(packet []byte) (uint32, []int16, error) {
	if len(packet) < HeaderSize {
		return 0, nil, ErrPacketTooSmall
	}

	sequence := binary.LittleEndian.Uint32(packet[:HeaderSize])
	payload := packet[HeaderSize:]
	samples := make([]int16, len(payload)/2)

	for i := 0; i < len(samples); i++ {
		offset := i * 2
		samples[i] = int16(binary.LittleEndian.Uint16(payload[offset : offset+2]))
	}

	return sequence, samples, nil
}

func SequenceAhead(sequence uint32, last uint32) bool {
	return int32(sequence-last) > 0
}
