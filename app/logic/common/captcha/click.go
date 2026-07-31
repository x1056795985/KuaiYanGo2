package captcha

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strconv"
	"strings"
	"sync"

	drawx "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

const (
	clickColumns       = 7
	clickRows          = 3
	iconSize           = 48
	clickCount         = 4
	maxClickDifficulty = clickColumns*clickRows - clickCount - 1
	imageDataPrefix    = "data:image/png;base64,"
)

var (
	//go:embed icon/*.png
	iconFiles embed.FS

	assetsOnce sync.Once
	assets     clickAssets
	assetsErr  error
	bufferPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
	pngEncoder = png.Encoder{
		CompressionLevel: png.BestSpeed,
		BufferPool:       &captchaPNGBufferPool,
	}
	captchaPNGBufferPool pngBufferPool
)

type pngBufferPool struct {
	pool sync.Pool
}

func (p *pngBufferPool) Get() *png.EncoderBuffer {
	if value := p.pool.Get(); value != nil {
		return value.(*png.EncoderBuffer)
	}
	return new(png.EncoderBuffer)
}

func (p *pngBufferPool) Put(buffer *png.EncoderBuffer) {
	p.pool.Put(buffer)
}

type clickAssets struct {
	prompt   image.Image
	blank    image.Image
	original [91]image.Image
	variants [91][]image.Image
}

// GenerateClick creates a click-in-order challenge. Difficulty controls the
// number of distractors and is clamped to the supported range.
func GenerateClick(difficulty int) (id, dataURL string, err error) {
	assetsOnce.Do(loadClickAssets)
	if assetsErr != nil {
		return "", "", assetsErr
	}
	if difficulty < 0 {
		difficulty = 0
	} else if difficulty > maxClickDifficulty {
		difficulty = maxClickDifficulty
	}

	id, err = randomToken(18, "23456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ")
	if err != nil {
		return "", "", err
	}

	icons := make([]int, 90)
	for i := range icons {
		icons[i] = i + 1
	}
	if err = secureShuffle(icons); err != nil {
		return "", "", err
	}
	icons = icons[:clickCount+difficulty]
	selected := append([]int(nil), icons[:clickCount]...)
	if err = secureShuffle(icons); err != nil {
		return "", "", err
	}

	grid := make([]int, clickColumns*clickRows)
	leftPadding := (len(grid) - len(icons)) / 2
	copy(grid[leftPadding:], icons)

	background := image.NewRGBA(image.Rect(0, 0, clickColumns*iconSize, (clickRows+1)*iconSize))
	draw.Draw(background, image.Rect(0, 0, 150, iconSize), assets.prompt, image.Point{}, draw.Src)
	for i, iconID := range selected {
		x := 150 + i*iconSize
		draw.Draw(background, image.Rect(x, 0, x+iconSize, iconSize), assets.original[iconID], image.Point{}, draw.Over)
	}
	draw.Draw(background, image.Rect(0, iconSize-1, background.Bounds().Dx(), iconSize), image.NewUniform(color.Black), image.Point{}, draw.Src)

	targets := make([]image.Rectangle, clickCount)
	for index, iconID := range grid {
		x := index % clickColumns * iconSize
		y := index/clickColumns*iconSize + iconSize
		rect := image.Rect(x, y, x+iconSize, y+iconSize)
		icon := assets.blank
		if iconID != 0 {
			variant, randomErr := randomInt(len(assets.variants[iconID]))
			if randomErr != nil {
				return "", "", randomErr
			}
			icon = assets.variants[iconID][variant]
			for selectedIndex, selectedID := range selected {
				if selectedID == iconID {
					targets[selectedIndex] = rect
				}
			}
		}
		draw.Draw(background, rect, icon, image.Point{}, draw.Over)
	}

	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)
	if err = pngEncoder.Encode(buffer, background); err != nil {
		return "", "", err
	}
	VerificationCodes.set(id, targets)
	return id, imageDataPrefix + base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// VerifyClick validates four ordered x|y points and optionally consumes the challenge.
func VerifyClick(id, answer string, consume bool) bool {
	if id == "" || answer == "" {
		return false
	}
	unlock := VerificationCodes.lock(id)
	defer unlock()

	raw, ok := VerificationCodes.getLocked(id)
	if !ok {
		return false
	}
	targets, ok := raw.([]image.Rectangle)
	if !ok || len(targets) != clickCount {
		return false
	}
	points := strings.Split(answer, ",")
	if len(points) != clickCount {
		return false
	}
	for i, encoded := range points {
		xText, yText, ok := strings.Cut(encoded, "|")
		if !ok {
			return false
		}
		x, xErr := strconv.Atoi(xText)
		y, yErr := strconv.Atoi(yText)
		if xErr != nil || yErr != nil || !image.Pt(x, y).In(targets[i]) {
			return false
		}
	}
	if consume {
		VerificationCodes.deleteLocked(id)
	}
	return true
}

func loadClickAssets() {
	assets.prompt, assetsErr = decodeEmbeddedImage("icon/请依次点击.png")
	if assetsErr != nil {
		return
	}
	assets.blank, assetsErr = decodeEmbeddedImage("icon/0.png")
	if assetsErr != nil {
		return
	}
	angles := [...]float64{-36, -24, -12, 0, 12, 24, 36}
	for iconID := 1; iconID <= 90; iconID++ {
		var source image.Image
		source, assetsErr = decodeEmbeddedImage(fmt.Sprintf("icon/%d.png", iconID))
		if assetsErr != nil {
			return
		}
		assets.original[iconID] = source
		distorted := distort(source, 10, 0.05)
		variants := make([]image.Image, len(angles))
		for i, angle := range angles {
			variants[i] = rotate(distorted, angle)
		}
		assets.variants[iconID] = variants
	}
}

func decodeEmbeddedImage(name string) (image.Image, error) {
	data, err := iconFiles.ReadFile(name)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode captcha asset %s: %w", name, err)
	}
	return img, nil
}

func secureShuffle[T any](values []T) error {
	for i := len(values) - 1; i > 0; i-- {
		j, err := randomInt(i + 1)
		if err != nil {
			return err
		}
		values[i], values[j] = values[j], values[i]
	}
	return nil
}

func rotate(source image.Image, angle float64) image.Image {
	radians := angle * math.Pi / 180
	center := float64(iconSize) / 2
	cosine, sine := math.Cos(radians), math.Sin(radians)
	matrix := f64.Aff3{
		cosine, -sine, center - cosine*center + sine*center,
		sine, cosine, center - sine*center - cosine*center,
	}
	destination := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))
	drawx.BiLinear.Transform(destination, matrix, source, source.Bounds(), draw.Over, nil)
	return destination
}

func distort(source image.Image, amplitude, frequency float64) image.Image {
	bounds := source.Bounds()
	destination := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		offset := int(amplitude * math.Sin(float64(y)*frequency))
		for x := 0; x < bounds.Dx(); x++ {
			sourceX := x + offset
			if sourceX >= 0 && sourceX < bounds.Dx() {
				destination.Set(x, y, source.At(bounds.Min.X+sourceX, bounds.Min.Y+y))
			}
		}
	}
	return destination
}
