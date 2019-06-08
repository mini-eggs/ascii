package handler

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/thanhpk/randstr"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

const (
	ASCIISTR = "MND8OZ$7I?+=~:,.."

	maxW = 200
	minW = 80
	defW = 100
)

/* route (the only one) */

/* GET: / */
func root(c *gin.Context) {
	fn := "/tmp/" + randstr.Hex(16)

	/* save image locally */

	width, err := strconv.Atoi(c.Query("w"))
	if err != nil {
		width = defW
	} else if width < minW {
		width = minW
	} else if width > maxW {
		width = maxW
	}

	img := c.Query("u")

	fmt.Printf("IMG : %s\n", img)

	if img == "" {
		/* fallback */
		img = "https://raw.githubusercontent.com/mini-eggs/Wigly/master/assets/500x.png"
	}

	client := http.Client{}
	resp, err := client.Get(img)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = ioutil.WriteFile(fn, body, 0644)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	/* now open and accii-ify */

	file, err := os.Open(fn)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	item, _, err := image.Decode(file)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	defer file.Close()

	c.String(http.StatusOK, string(ascii(scale(item, width))))
}

/* helpers */

func scale(img image.Image, w int) (image.Image, int, int) {
	bounds := img.Bounds()
	h := (bounds.Max.Y * w * 10) / (bounds.Max.X * 16)
	img = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
	return img, w, h
}

func ascii(img image.Image, w, h int) []byte {
	table := []byte(ASCIISTR)
	buf := new(bytes.Buffer)

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			g := color.GrayModel.Convert(img.At(j, i))
			y := reflect.ValueOf(g).FieldByName("Y").Uint()
			pos := int(y * 16 / 255)

			_ = buf.WriteByte(table[pos])
		}

		_ = buf.WriteByte('\n')
	}

	return buf.Bytes()
}

/* main */

func build() *gin.Engine {
	r := gin.Default()
	r.GET("/", root)
	return r
}

/* for local dev */
func main() {
	build().Run()
}

/* for zeit */
func H(w http.ResponseWriter, r *http.Request) {
	build().ServeHTTP(w, r)
}
