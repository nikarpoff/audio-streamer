package protocol

import (
	"encoding/binary"
	"errors"

	"github.com/nikarpoff/audio-streamer/internal/audio"
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

// EncodeAudioPacketMono encodes input interleaved PCM to mono payload by averaging
// channels per frame before writing packet data.
func EncodeAudioPacketMono(sequence uint32, pcm []int16, inputChannels int) []byte {
	mono := audio.MixToMonoInterleaved(pcm, inputChannels)
	return EncodeAudioPacket(sequence, mono)
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

// DecodeAudioPacketToChannels decodes mono payload and duplicates it into
// interleaved output channel count for playback.
func DecodeAudioPacketToChannels(packet []byte, outputChannels int) (uint32, []int16, error) {
	sequence, mono, err := DecodeAudioPacket(packet)
	if err != nil {
		return 0, nil, err
	}

	return sequence, audio.MonoToInterleaved(mono, outputChannels), nil
}

func SequenceAhead(sequence uint32, last uint32) bool {
	return int32(sequence-last) > 0
}
