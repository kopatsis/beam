package reviewhelp

import (
	"beam/data/models"
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

func ProcessImages(imgs []models.IntermImage) []models.IntermImage {
	if len(imgs) > 3 {
		imgs = imgs[:3]
	}

	var processedImgs []models.IntermImage

	for _, img := range imgs {
		if img.FileType != "image/png" && img.FileType != "image/jpeg" {
			continue
		} else if len(img.Data) == 0 {
			continue
		}

		var scaleFactor float64
		switch {
		case len(img.Data) > 1000*1024:
			scaleFactor = 0.5
		case len(img.Data) > 500*1024:
			scaleFactor = 0.65
		case len(img.Data) > 200*1024:
			scaleFactor = 0.8
		case len(img.Data) > 100*1024:
			scaleFactor = 0.9
		default:
			scaleFactor = 1
		}

		if scaleFactor < 1 {
			img.Data = DownscaleImage(img.Data, scaleFactor)
		}

		img.AddedID = "img-" + uuid.NewString()
		if len(img.FileNameOG) > 512 {
			img.FileNameNew = img.AddedID + img.FileNameOG[len(img.FileNameOG)-512:]
		} else {
			img.FileNameNew = img.AddedID + img.FileNameOG
		}

		processedImgs = append(processedImgs, img)
	}

	return processedImgs
}

func ProcessImagesSingle(img models.IntermImage) models.IntermImage {
	procList := ProcessImages([]models.IntermImage{img})
	if len(procList) > 0 {
		return procList[0]
	}
	return models.IntermImage{}
}

func DownscaleImage(data []byte, scaleFactor float64) []byte {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}

	newWidth := uint(float64(img.Bounds().Dx()) * scaleFactor)
	newHeight := uint(float64(img.Bounds().Dy()) * scaleFactor)

	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	var buf bytes.Buffer
	switch {
	case data[0] == 0x89 && data[1] == 0x50:
		err = png.Encode(&buf, resizedImg)
	case data[0] == 0xFF && data[1] == 0xD8:
		err = jpeg.Encode(&buf, resizedImg, nil)
	default:
		return data
	}

	if err != nil {
		return data
	}

	return buf.Bytes()
}

func GetImgsFromReq(c *gin.Context) ([]models.IntermImage, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := form.File["images"]
	var images []models.IntermImage

	for _, fileHeader := range files {

		if len(images) >= 3 {
			break
		}

		if fileHeader.Size > 5*512*1024 {
			continue
		}

		fileType := fileHeader.Header.Get("Content-Type")
		if fileType != "image/png" && fileType != "image/jpeg" {
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		image := models.IntermImage{
			Data:       data,
			FileNameOG: fileHeader.Filename,
			FileType:   fileType,
		}

		images = append(images, image)
	}

	return images, nil
}

func GetImgSingleFromReq(c *gin.Context) (*models.IntermImage, error) {
	imgs, err := GetImgsFromReq(c)
	if err != nil {
		return nil, err
	}

	if len(imgs) > 0 {
		return &imgs[0], nil
	}
	return nil, nil
}
