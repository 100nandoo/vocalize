package audio

import (
	"bytes"
	"encoding/binary"
	"os"

	"github.com/hraban/opus"
)

// oggCRC computes the Ogg-specific CRC32 (non-reflected, poly 0x04C11DB7).
func oggCRC(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc ^= uint32(b) << 24
		for i := 0; i < 8; i++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ 0x04C11DB7
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func buildOggPage(headerType byte, granule int64, serial, seqno uint32, data []byte) []byte {
	remaining := len(data)
	var segs []byte
	for remaining > 0 {
		seg := remaining
		if seg > 255 {
			seg = 255
		}
		segs = append(segs, byte(seg))
		remaining -= seg
	}
	if len(segs) == 0 || segs[len(segs)-1] == 255 {
		segs = append(segs, 0)
	}

	var page bytes.Buffer
	page.WriteString("OggS")
	page.WriteByte(0)
	page.WriteByte(headerType)
	binary.Write(&page, binary.LittleEndian, granule)
	binary.Write(&page, binary.LittleEndian, serial)
	binary.Write(&page, binary.LittleEndian, seqno)
	binary.Write(&page, binary.LittleEndian, uint32(0)) // CRC placeholder
	page.WriteByte(byte(len(segs)))
	page.Write(segs)
	page.Write(data)

	b := page.Bytes()
	binary.LittleEndian.PutUint32(b[22:26], oggCRC(b))
	return b
}

// EncodePCMToOpus encodes 16-bit LE mono PCM into an Ogg Opus file.
func EncodePCMToOpus(pcm []byte, sampleRate int) ([]byte, error) {
	enc, err := opus.NewEncoder(sampleRate, 1, opus.AppVoIP)
	if err != nil {
		return nil, err
	}

	frameSize := sampleRate / 50 // 20ms frames

	samples := make([]int16, len(pcm)/2)
	for i := range samples {
		samples[i] = int16(pcm[2*i]) | int16(pcm[2*i+1])<<8
	}

	var frames [][]byte
	buf := make([]byte, 4000)
	for i := 0; i+frameSize <= len(samples); i += frameSize {
		n, err := enc.Encode(samples[i:i+frameSize], buf)
		if err != nil {
			return nil, err
		}
		packet := make([]byte, n)
		copy(packet, buf[:n])
		frames = append(frames, packet)
	}

	const serial = uint32(0x12345678)
	var out bytes.Buffer
	var seqno uint32

	// OpusHead identification header (RFC 7845)
	var head bytes.Buffer
	head.WriteString("OpusHead")
	head.WriteByte(1) // version
	head.WriteByte(1) // channels
	binary.Write(&head, binary.LittleEndian, uint16(0))          // pre-skip
	binary.Write(&head, binary.LittleEndian, uint32(sampleRate)) // input sample rate
	binary.Write(&head, binary.LittleEndian, uint16(0))          // output gain
	head.WriteByte(0)                                             // mapping family (0 = mono/stereo)
	out.Write(buildOggPage(0x02, 0, serial, seqno, head.Bytes()))
	seqno++

	// OpusTags comment header
	var tags bytes.Buffer
	tags.WriteString("OpusTags")
	vendor := "inti"
	binary.Write(&tags, binary.LittleEndian, uint32(len(vendor)))
	tags.WriteString(vendor)
	binary.Write(&tags, binary.LittleEndian, uint32(0)) // no user comments
	out.Write(buildOggPage(0x00, 0, serial, seqno, tags.Bytes()))
	seqno++

	// Audio pages — one Opus packet per page, granule in 48kHz units (RFC 7845)
	granulePerFrame := int64(frameSize * 48000 / sampleRate)
	var granule int64
	for i, frame := range frames {
		granule += granulePerFrame
		headerType := byte(0x00)
		if i == len(frames)-1 {
			headerType = 0x04 // EOS
		}
		out.Write(buildOggPage(headerType, granule, serial, seqno, frame))
		seqno++
	}

	return out.Bytes(), nil
}

func WriteOpusFile(path string, pcm []byte, sampleRate int) error {
	data, err := EncodePCMToOpus(pcm, sampleRate)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
