package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestGenerateAndVerifyClick(t *testing.T) {
	previousStore := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = previousStore })

	id, dataURL, err := GenerateClick(-1)
	if err != nil {
		t.Fatalf("GenerateClick() error = %v", err)
	}
	if len(id) != 18 {
		t.Fatalf("challenge id length = %d, want 18", len(id))
	}
	if !strings.HasPrefix(dataURL, imageDataPrefix) {
		t.Fatalf("image does not have data URL prefix")
	}
	imageBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, imageDataPrefix))
	if err != nil {
		t.Fatalf("image is not valid base64: %v", err)
	}
	generatedImage, err := png.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		t.Fatalf("image is not valid PNG: %v", err)
	}
	if got, want := generatedImage.Bounds(), image.Rect(0, 0, clickColumns*iconSize, (clickRows+1)*iconSize); got != want {
		t.Fatalf("image bounds = %v, want %v", got, want)
	}

	raw, ok := VerificationCodes.backend().Get(cachePrefix + id)
	if !ok {
		t.Fatal("challenge was not cached")
	}
	targets, ok := raw.([]image.Rectangle)
	if !ok || len(targets) != clickCount {
		t.Fatalf("cached targets = %#v", raw)
	}
	answerParts := make([]string, len(targets))
	for i, target := range targets {
		answerParts[i] = fmt.Sprintf("%d|%d", target.Min.X+1, target.Min.Y+1)
	}
	answer := strings.Join(answerParts, ",")
	if !VerifyClick(id, answer, true) {
		t.Fatal("correct click answer failed")
	}
	if VerifyClick(id, answer, true) {
		t.Fatal("consumed click answer passed twice")
	}
}

func TestVerifyClickRejectsMalformedAnswerWithoutConsuming(t *testing.T) {
	previousStore := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = previousStore })

	VerificationCodes.set("id", []image.Rectangle{
		image.Rect(0, 0, 48, 48),
		image.Rect(48, 0, 96, 48),
		image.Rect(96, 0, 144, 48),
		image.Rect(144, 0, 192, 48),
	})
	if VerifyClick("id", "bad", true) {
		t.Fatal("malformed click answer passed")
	}
	if !VerifyClick("id", "1|1,49|1,97|1,145|1", true) {
		t.Fatal("valid answer failed after malformed attempt")
	}
}

func BenchmarkGenerateClick(b *testing.B) {
	previousStore := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	b.Cleanup(func() { VerificationCodes = previousStore })
	_, _, _ = GenerateClick(8)
	b.ResetTimer()
	for range b.N {
		if _, _, err := GenerateClick(8); err != nil {
			b.Fatal(err)
		}
	}
}
