package audio

// MixToMonoInterleaved mixes interleaved PCM samples into mono by averaging
// all channels per frame.
func MixToMonoInterleaved(samples []int16, channels int) []int16 {
	if channels <= 1 {
		out := make([]int16, len(samples))
		copy(out, samples)
		return out
	}

	frames := len(samples) / channels
	out := make([]int16, frames)
	for frame := 0; frame < frames; frame++ {
		base := frame * channels
		var sum int32
		for ch := 0; ch < channels; ch++ {
			sum += int32(samples[base+ch])
		}
		out[frame] = int16(sum / int32(channels))
	}

	return out
}

// MonoToInterleaved duplicates mono samples to requested interleaved channel count.
func MonoToInterleaved(mono []int16, channels int) []int16 {
	if channels <= 1 {
		out := make([]int16, len(mono))
		copy(out, mono)
		return out
	}

	out := make([]int16, len(mono)*channels)
	for i, sample := range mono {
		base := i * channels
		for ch := 0; ch < channels; ch++ {
			out[base+ch] = sample
		}
	}

	return out
}
