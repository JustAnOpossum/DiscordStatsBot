package main

import (
	_ "image/jpeg"
	_ "image/png"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// type returnedColors struct {
// 	Main      vibrant.Color
// 	Secondary vibrant.Color
// }

// var pixelWidthBetween = [5]float64{0, 305, 205, 155, 125}
// var pixelWidthStart = [5]float64{620, 475, 420, 395, 380}

// func getColorPallete(img *image.Image) (returnedColors, error) {
// 	pallete, err := vibrant.NewPaletteFromImage(*img)
// 	if err != nil {
// 		errors.Wrap(err, "Generating Color Pallete")
// 	}
// 	colors := pallete.ExtractAwesome()

// 	if colors["Vibrant"] == nil || colors["DarkMuted"] == nil {
// 		return getMissingColors(colors), nil
// 	}
// 	return returnedColors{
// 		Main:      colors["Vibrant"].Color,
// 		Secondary: colors["DarkMuted"].Color,
// 	}, nil
// }

// func getMissingColors(colors map[string]*vibrant.Swatch) returnedColors {
// 	switch len(colors) {
// 	case 0:
// 		return returnedColors{}
// 	case 1:
// 		for _, swatch := range colors {
// 			return returnedColors{
// 				Secondary: swatch.Color,
// 			}
// 		}
// 	default:
// 		var names []string
// 		for name := range colors {
// 			names = append(names, name)
// 		}
// 		return returnedColors{
// 			Main:      colors[names[0]].Color,
// 			Secondary: colors[names[1]].Color,
// 		}
// 	}
// 	return returnedColors{}
// }

// func addCircleIcon(img *image.Image, base *imagick.MagickWand) error {
// 	profilePic := imagick.NewMagickWand()
// 	defer profilePic.Destroy()
// 	buffer := new(bytes.Buffer)
// 	err := png.Encode(buffer, *img)
// 	if err != nil {
// 		return errors.Wrap(err, "Loading Image into Wand")
// 	}
// 	profilePic.ReadImageBlob(buffer.Bytes())
// 	profilePic.ResizeImage(450, 450, imagick.FILTER_UNDEFINED)
// 	height := profilePic.GetImageHeight()
// 	width := profilePic.GetImageWidth()

// 	circleMask := imagick.NewMagickWand()
// 	pw := imagick.NewPixelWand()
// 	circleDraw := imagick.NewDrawingWand()
// 	defer pw.Destroy()
// 	defer circleMask.Destroy()
// 	defer circleMask.Destroy()
// 	pw.SetColor("black")
// 	circleMask.NewImage(height, width, pw)

// 	pw.SetColor("white")
// 	circleDraw.SetFillColor(pw)
// 	circleDraw.Circle(float64(height/2), float64(width/2), float64(height/2), 0)
// 	circleMask.DrawImage(circleDraw)

// 	circleMask.SetImageMatte(false)
// 	profilePic.SetImageMatte(false)
// 	profilePic.CompositeImage(circleMask, imagick.COMPOSITE_OP_COPY_ALPHA, true, 0, 0)

// 	base.CompositeImage(profilePic, imagick.COMPOSITE_OP_OVER, true, -90, -120)
// 	return nil
// }

// func drawText(base *imagick.MagickWand, name, hoursPlayed, gamesPlayed string, colors returnedColors) error {
// 	textWand := imagick.NewDrawingWand()
// 	textColor := imagick.NewPixelWand()
// 	defer textWand.Destroy()
// 	defer textColor.Destroy()
// 	textColor.SetColor(colors.Main.RGBHex())
// 	textWand.SetFont(path.Join(dataDir, "main.ttf"))
// 	textWand.SetFillColor(textColor)
// 	textWand.SetFontSize(70)
// 	textWand.SetGravity(imagick.GRAVITY_CENTER)

// 	if len(name) >= 16 {
// 		name = name[:16] + "\n" + name[16:]
// 	}

// 	textWand.Annotation(-345, 0, hoursPlayed+"\nHours\nPlayed")
// 	textWand.Annotation(-345, 310, gamesPlayed+"\nGames\nPlayed")

// 	textWand.SetFontSize(100)
// 	textColor.SetColor(colors.Secondary.RGBHex())
// 	textWand.SetFillColor(textColor)
// 	textWand.Annotation(170, -300, name+"\nStats")
// 	base.DrawImage(textWand)
// 	return nil
// }

// func drawBotText(base *imagick.MagickWand, name, totalStats, totalGames, totalImgGenerated, totalServers, totalUsers string, colors returnedColors) error {
// 	textWand := imagick.NewDrawingWand()
// 	textColor := imagick.NewPixelWand()
// 	defer textWand.Destroy()
// 	defer textColor.Destroy()
// 	textColor.SetColor(colors.Main.RGBHex())
// 	textWand.SetFont("main.ttf")
// 	textWand.SetFillColor(textColor)
// 	textWand.SetGravity(imagick.GRAVITY_CENTER)

// 	textWand.SetFontSize(70)
// 	textWand.Annotation(-345, 0, totalStats+"\nTotal\nStats")
// 	textWand.Annotation(-345, 310, totalGames+"\nTotal\nGames")
// 	textWand.Annotation(-40, 0, totalUsers+"\nTotal\nUsers")
// 	textWand.Annotation(-40, 310, totalServers+"\nTotal\nServers")
// 	textWand.SetFontSize(90)
// 	textWand.Annotation(290, 150, totalImgGenerated+"\nImages\nGenerated!")

// 	textWand.SetFontSize(100)
// 	textColor.SetColor(colors.Secondary.RGBHex())
// 	textWand.SetFillColor(textColor)
// 	textWand.Annotation(170, -300, name+"\nStats")

// 	base.DrawImage(textWand)
// 	return nil
// }

// func addGraph() error {
// 	return nil
// }

func main() {
	var err error

	imagick.Initialize()
	defer imagick.Terminate()
	mainImg := imagick.NewMagickWand()
	defer mainImg.Destroy()
	bgColor := imagick.NewPixelWand()
	defer bgColor.Destroy()

	// colors, _ := getColorPallete(img)
	bgColor.SetColor("#b5332a")
	mainImg.NewImage(1000, 1000, bgColor)

	// err = drawCircles(mainImg, colors, "mask.png")
	// err = addCircleIcon(img, mainImg)
	// err = drawText(mainImg, name, hoursPlayed, gamesPlayed, colors)
	// err = addGraph(mainImg, graphType, img, colors, userID)

	if err != nil {

	}

	mainImg.WriteImage("out.png")
}

func drawCircles(base *imagick.MagickWand, colors returnedColors, maskToUse string) error {
	mask := imagick.NewMagickWand()
	pw := imagick.NewPixelWand()
	defer mask.Destroy()
	defer pw.Destroy()
	mainImg := imagick.NewMagickWand()
	mask.ReadImage(maskToUse)

	pw.SetColor(colors.Secondary.RGBHex())
	mainImg.NewImage(1000, 1000, pw)

	mainImg.SetImageMatte(false)
	mask.SetImageMatte(false)

	mainImg.CompositeImage(mask, imagick.COMPOSITE_OP_COPY_ALPHA, true, 0, 0)
	base.CompositeImage(mainImg, imagick.COMPOSITE_OP_OVER, true, 0, 0)
	return nil
}
