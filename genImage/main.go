package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/generaltso/vibrant"
	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type returnedColors struct {
	Main      vibrant.Color
	Secondary vibrant.Color
}

type gameInfo struct {
	Name   string
	Icon   *image.Image
	Colors returnedColors
	Hours  float64
	Path   string
}

var pixelWidthBetween = [5]float64{0, 305, 205, 155, 125}
var pixelWidthStart = [5]float64{620, 475, 420, 395, 380}

//Environment vars have dir names
//data is tructed like this
//1: Total Hours
//2: Total Games
//3: Name
//4: Graph Type
//5: avatarURL
//6-10 Game names
func main() {
	var err error

	//Gets all the data it needs to function
	staticDir := os.Getenv("STATIC_DIR")
	imageDir := os.Getenv("IMAGE_DIR")
	info, _ := ioutil.ReadFile(path.Join(imageDir, "data"))
	params := strings.Split(string(info), "\n")
	avatarURL := params[4]
	games := make([]gameInfo, 0)
	//This loops gets all the info for the games and adds it to a object that can be passed later
	for i := 5; i < len(params); i++ {
		imageFile, err := os.Open(path.Join(imageDir, params[i], "icon"))
		if err != nil {
			panic(err)
		}
		defer imageFile.Close()
		image, _, err := image.Decode(imageFile)
		if err != nil {
			panic(err)
		}
		hoursFile, _ := ioutil.ReadFile(path.Join(imageDir, params[i], "hours"))
		hours, _ := strconv.ParseFloat(string(hoursFile), 64)
		colors, _ := getColorPallete(&image)
		info := gameInfo{
			Name:   params[i],
			Icon:   &image,
			Hours:  hours,
			Colors: colors,
			Path:   path.Join(imageDir, params[i], "icon"),
		}
		games = append(games, info)
	}
	//Loads the avatar image, if there is an error then the image will be the error image
	avatarImage := loadDiscordAvatar(avatarURL)
	if avatarImage == nil {
		errorFile, _ := os.Open(path.Join(staticDir, "avatarError.png"))
		defer errorFile.Close()
		newImg, _, _ := image.Decode(errorFile)
		avatarImage = &newImg
	}
	hoursPlayed := params[0]
	gamesPlayed := params[1]
	name := params[2]
	graphType := params[3]

	imagick.Initialize()
	defer imagick.Terminate()
	mainImg := imagick.NewMagickWand()
	defer mainImg.Destroy()
	bgColor := imagick.NewPixelWand()
	defer bgColor.Destroy()

	colors, _ := getColorPallete(avatarImage)
	bgColor.SetColor(colors.Main.RGBHex())
	mainImg.NewImage(1000, 1000, bgColor)

	err = drawCircles(mainImg, colors, path.Join(staticDir, "mask.png"))
	err = addCircleIcon(avatarImage, mainImg)
	err = drawText(mainImg, name, hoursPlayed, gamesPlayed, colors, path.Join(staticDir, "main.ttf"))
	err = addGraph(mainImg, graphType, games, colors)

	if err != nil {
		panic(err)
	}
	mainImg.WriteImage(path.Join(imageDir, "out.png"))
}

func addGraph(base *imagick.MagickWand, graphType string, games []gameInfo, colors returnedColors) error {
	graphWand := imagick.NewMagickWand()
	iconWand := imagick.NewMagickWand()
	iconDrawingWand := imagick.NewDrawingWand()
	defer graphWand.Destroy()
	defer iconWand.Destroy()
	defer iconDrawingWand.Destroy()

	switch graphType {
	case "bar":
		var lengthOfGraph int
		if len(games) >= 5 {
			lengthOfGraph = 4
		} else {
			lengthOfGraph = len(games) - 1
		}

		for i := range games {
			iconWand.ReadImage(games[i].Path)
			iconWand.ResizeImage(100, 100, imagick.FILTER_UNDEFINED)
			var whereToDraw float64
			if i == 0 {
				whereToDraw = pixelWidthStart[lengthOfGraph]
			} else {
				whereToDraw = pixelWidthStart[lengthOfGraph] + (pixelWidthBetween[lengthOfGraph] * float64(i))
			}
			iconDrawingWand.Composite(imagick.COMPOSITE_OP_OVER, whereToDraw, 850, 100, 100, iconWand)
		}
		base.DrawImage(iconDrawingWand)
		graph, err := createBarChart(games, colors)
		if err != nil {
			return err
		}
		graphWand.ReadImageBlob(graph.Bytes())
		base.CompositeImage(graphWand, imagick.COMPOSITE_OP_OVER, true, 350, 350)
		break

	case "pie":
		pieChart, err := createPieChart(games)
		if err != nil {
			return err
		}
		graphWand.ReadImageBlob(pieChart.Bytes())
		base.CompositeImage(graphWand, imagick.COMPOSITE_OP_OVER, true, 350, 350)
		break
	}
	return nil
}

func getColorPallete(img *image.Image) (returnedColors, error) {
	pallete, err := vibrant.NewPaletteFromImage(*img)
	if err != nil {
		errors.Wrap(err, "Generating Color Pallete")
	}
	colors := pallete.ExtractAwesome()

	if colors["Vibrant"] == nil || colors["DarkMuted"] == nil {
		return getMissingColors(colors), nil
	}
	return returnedColors{
		Main:      colors["Vibrant"].Color,
		Secondary: colors["DarkMuted"].Color,
	}, nil
}

func getMissingColors(colors map[string]*vibrant.Swatch) returnedColors {
	switch len(colors) {
	case 0:
		return returnedColors{}
	case 1:
		for _, swatch := range colors {
			return returnedColors{
				Secondary: swatch.Color,
			}
		}
	default:
		var names []string
		for name := range colors {
			names = append(names, name)
		}
		return returnedColors{
			Main:      colors[names[0]].Color,
			Secondary: colors[names[1]].Color,
		}
	}
	return returnedColors{}
}

func loadDiscordAvatar(URL string) *image.Image {
	//Sets up inital checks to see if the data is valid
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Get(URL)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil
	}

	img, _, err := image.Decode(res.Body)
	if err != nil {
		return nil
	}
	return &img
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

func addCircleIcon(img *image.Image, base *imagick.MagickWand) error {
	profilePic := imagick.NewMagickWand()
	defer profilePic.Destroy()
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, *img)
	if err != nil {
		return errors.Wrap(err, "Loading Image into Wand")
	}
	profilePic.ReadImageBlob(buffer.Bytes())
	profilePic.ResizeImage(450, 450, imagick.FILTER_UNDEFINED)
	height := profilePic.GetImageHeight()
	width := profilePic.GetImageWidth()

	circleMask := imagick.NewMagickWand()
	pw := imagick.NewPixelWand()
	circleDraw := imagick.NewDrawingWand()
	defer pw.Destroy()
	defer circleMask.Destroy()
	defer circleMask.Destroy()
	pw.SetColor("black")
	circleMask.NewImage(height, width, pw)

	pw.SetColor("white")
	circleDraw.SetFillColor(pw)
	circleDraw.Circle(float64(height/2), float64(width/2), float64(height/2), 0)
	circleMask.DrawImage(circleDraw)

	circleMask.SetImageMatte(false)
	profilePic.SetImageMatte(false)
	profilePic.CompositeImage(circleMask, imagick.COMPOSITE_OP_COPY_ALPHA, true, 0, 0)

	base.CompositeImage(profilePic, imagick.COMPOSITE_OP_OVER, true, -90, -120)
	return nil
}

func drawText(base *imagick.MagickWand, name, hoursPlayed, gamesPlayed string, colors returnedColors, fontDir string) error {
	textWand := imagick.NewDrawingWand()
	textColor := imagick.NewPixelWand()
	defer textWand.Destroy()
	defer textColor.Destroy()
	textColor.SetColor(colors.Main.RGBHex())
	textWand.SetFont(fontDir)
	textWand.SetFillColor(textColor)
	textWand.SetFontSize(70)
	textWand.SetGravity(imagick.GRAVITY_CENTER)

	if len(name) >= 16 {
		name = name[:16] + "\n" + name[16:]
	}

	textWand.Annotation(-345, 0, hoursPlayed+"\nHours\nPlayed")
	textWand.Annotation(-345, 310, gamesPlayed+"\nGames\nPlayed")

	textWand.SetFontSize(100)
	textColor.SetColor(colors.Secondary.RGBHex())
	textWand.SetFillColor(textColor)
	textWand.Annotation(170, -300, name+"\nStats")
	base.DrawImage(textWand)
	return nil
}
