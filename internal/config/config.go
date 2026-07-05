package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey     string
	SpeechProvider   string
	KokoroHeartURL   string
	DefaultVoice     string
	DefaultModel     string
	Port             int
	Host             string
	MainKey          string
	TelegramBotToken string

	SummarizerProvider string // "gemini" | "groq" | "openrouter"
	GroqAPIKey         string
	GroqModel          string
	OpenRouterAPIKey   string
	OpenRouterModel    string
}

const (
	SpeechProviderGemini      = "gemini"
	SpeechProviderKokoroHeart = "kokoro-heart"
	DefaultKokoroHeartURL     = "https://koboldai-koboldcpp-tiefighter.hf.space/v1/audio/speech"
)

var validVoices = []string{
	"Zephyr", "Puck", "Charon", "Kore", "Fenrir",
	"Leda", "Orus", "Aoede", "Callirrhoe", "Autonoe",
	"Enceladus", "Iapetus", "Umbriel", "Algieba", "Despina",
	"Erinome", "Algenib", "Rasalgethi", "Laomedeia", "Achernar",
	"Alnilam", "Schedar", "Gacrux", "Pulcherrima", "Achird",
	"Zubenelgenubi", "Vindemiatrix", "Sadachbia", "Sadaltager", "Sulafat",
}

var kokoroHeartVoices = []string{"cheery"}

var validModels = []string{
	"gemini-2.5-flash-preview-tts",
	"gemini-2.5-pro-preview-tts",
	"gemini-3.1-flash-tts-preview",
}

const DefaultModelName = "gemini-3.1-flash-tts-preview"
const DefaultGeminiVoice = "Kore"
const DefaultKokoroHeartVoice = "cheery"

var placeholderSecretValues = map[string]struct{}{
	"":                 {},
	"put_api_key_here": {},
	"your_api_key":     {},
	"changeme":         {},
	"change_me":        {},
}

func ValidVoices() []string { return validVoices }

func ValidVoicesForProvider(provider string) []string {
	switch provider {
	case SpeechProviderGemini:
		return append([]string(nil), validVoices...)
	case SpeechProviderKokoroHeart:
		return append([]string(nil), kokoroHeartVoices...)
	default:
		return nil
	}
}

func IsValidVoice(name string) bool {
	for _, v := range validVoices {
		if v == name {
			return true
		}
	}
	return false
}

func IsValidVoiceForProvider(provider, name string) bool {
	for _, v := range ValidVoicesForProvider(provider) {
		if v == name {
			return true
		}
	}
	return false
}

func ValidModels() []string { return validModels }

func ValidModelsForProvider(provider string) []string {
	switch provider {
	case SpeechProviderGemini:
		return append([]string(nil), validModels...)
	case SpeechProviderKokoroHeart:
		return []string{}
	default:
		return nil
	}
}

func IsValidModel(name string) bool {
	for _, m := range validModels {
		if m == name {
			return true
		}
	}
	return false
}

func IsValidModelForProvider(provider, name string) bool {
	if provider == SpeechProviderKokoroHeart {
		return name == ""
	}
	for _, m := range ValidModelsForProvider(provider) {
		if m == name {
			return true
		}
	}
	return false
}

func DefaultVoiceForProvider(provider string) string {
	switch provider {
	case SpeechProviderKokoroHeart:
		return DefaultKokoroHeartVoice
	case SpeechProviderGemini:
		fallthrough
	default:
		return DefaultGeminiVoice
	}
}

func DefaultModelForProvider(provider string) string {
	switch provider {
	case SpeechProviderKokoroHeart:
		return ""
	case SpeechProviderGemini:
		fallthrough
	default:
		return DefaultModelName
	}
}

func ValidSpeechProviders() []string {
	return []string{SpeechProviderGemini, SpeechProviderKokoroHeart}
}

func IsValidSpeechProvider(provider string) bool {
	switch provider {
	case SpeechProviderGemini, SpeechProviderKokoroHeart:
		return true
	default:
		return false
	}
}

func loadSecret(names ...string) string {
	for _, name := range names {
		value := normalizeSecret(os.Getenv(name))
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeSecret(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	if _, isPlaceholder := placeholderSecretValues[strings.ToLower(value)]; isPlaceholder {
		return ""
	}
	return value
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	key := loadSecret("GEMINI_API_KEY", "GOOGLE_API_KEY")

	port := 8282
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	host := "127.0.0.1"
	if h := os.Getenv("HOST"); h != "" {
		host = h
	}

	speechProvider := strings.TrimSpace(os.Getenv("SPEECH_PROVIDER"))
	if !IsValidSpeechProvider(speechProvider) {
		speechProvider = SpeechProviderGemini
	}

	voice := DefaultVoiceForProvider(speechProvider)
	if v := os.Getenv("DEFAULT_VOICE"); v != "" && IsValidVoiceForProvider(speechProvider, v) {
		voice = v
	}

	model := DefaultModelForProvider(speechProvider)
	if m := os.Getenv("DEFAULT_MODEL"); m != "" && IsValidModelForProvider(speechProvider, m) {
		model = m
	}

	kokoroHeartURL := strings.TrimSpace(os.Getenv("KOKORO_HEART_URL"))
	if kokoroHeartURL == "" {
		kokoroHeartURL = DefaultKokoroHeartURL
	}

	// Auto-detect summarizer provider if not explicitly set.
	summarizerProvider := os.Getenv("SUMMARIZER_PROVIDER")
	if summarizerProvider == "" {
		switch {
		case key != "":
			summarizerProvider = "gemini"
		case loadSecret("GROQ_API_KEY") != "":
			summarizerProvider = "groq"
		case loadSecret("OPENROUTER_API_KEY") != "":
			summarizerProvider = "openrouter"
		}
	}

	groqModel := "openai/gpt-oss-120b"
	if m := os.Getenv("GROQ_MODEL"); m != "" {
		groqModel = m
	}

	openRouterModel := "openrouter/free"
	if m := os.Getenv("OPENROUTER_MODEL"); m != "" {
		openRouterModel = m
	}

	mainKey := os.Getenv("INTI_MAIN_KEY")
	if mainKey == "" {
		mainKey = os.Getenv("INTI_MASTER_KEY")
	}

	return &Config{
		GeminiAPIKey:       key,
		SpeechProvider:     speechProvider,
		KokoroHeartURL:     kokoroHeartURL,
		DefaultVoice:       voice,
		DefaultModel:       model,
		Port:               port,
		Host:               host,
		MainKey:            mainKey,
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		SummarizerProvider: summarizerProvider,
		GroqAPIKey:         loadSecret("GROQ_API_KEY"),
		GroqModel:          groqModel,
		OpenRouterAPIKey:   loadSecret("OPENROUTER_API_KEY"),
		OpenRouterModel:    openRouterModel,
	}, nil
}
