package audio

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

const sampleRate = 24000

// Play writes PCM to a temp Opus file and plays it via the platform audio player.
func Play(pcm []byte) error {
	player, args, err := platformPlayer()
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "inti-*.opus")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	opusData, err := EncodePCMToOpus(pcm, sampleRate)
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("encode opus: %w", err)
	}
	if _, err := tmp.Write(opusData); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write opus: %w", err)
	}
	tmp.Close()

	cmdArgs := append(args, tmpPath)
	if err := exec.Command(player, cmdArgs...).Run(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("play audio: %w", err)
	}

	os.Remove(tmpPath)
	return nil
}

func platformPlayer() (string, []string, error) {
	// Prefer players with native Opus support
	for _, p := range []string{"mpv", "ffplay", "cvlc"} {
		if _, err := exec.LookPath(p); err == nil {
			if p == "ffplay" {
				return p, []string{"-nodisp", "-autoexit"}, nil
			}
			if p == "cvlc" {
				return p, []string{"--play-and-exit"}, nil
			}
			return p, nil, nil
		}
	}
	switch runtime.GOOS {
	case "darwin":
		return "", nil, fmt.Errorf("no Opus-capable player found (install mpv: brew install mpv)")
	case "linux":
		return "", nil, fmt.Errorf("no Opus-capable player found (install mpv, ffplay, or vlc)")
	default:
		return "", nil, fmt.Errorf("no Opus-capable player found (install mpv or ffplay)")
	}
}
