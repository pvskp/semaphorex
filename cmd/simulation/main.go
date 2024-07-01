package main

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	bg        *ebiten.Image
	redCar    *ebiten.Image
	blueCar   *ebiten.Image
	purpleCar *ebiten.Image
	yellowCar *ebiten.Image

	serverConn    *grpc.ClientConn
	serverClient  pb.VehicleDiscoveryClient
	serverAddress string

	upSlice    = []*pb.Vehicle{}
	downSlice  = []*pb.Vehicle{}
	rightSlice = []*pb.Vehicle{}
	leftSlice  = []*pb.Vehicle{}
)

const (
	width  = 600
	height = 600
)

// Game implements ebiten.Game interface.
type Game struct{}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	checkForUpdates()
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	bgOp := &ebiten.DrawImageOptions{}
	screen.DrawImage(bg, bgOp)

	for _, vehicle := range vehicleList {
		spawnVehicle(screen, vehicle)
	}

	spawnRedCar(screen)
	spawnBlueCar(screen)
	spawnYellowCar(screen)
	spawnPurpleCar(screen)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func main() {
	game := &Game{}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("Your game's title")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func spawnRedCar(screen *ebiten.Image) {
	redCarOp := &ebiten.DrawImageOptions{}
	redCarOp.GeoM.Rotate(math.Pi / 2)
	redCarOp.GeoM.Translate(280, 0)
	screen.DrawImage(redCar, redCarOp)
}

func spawnBlueCar(screen *ebiten.Image) {
	blueCarOp := &ebiten.DrawImageOptions{}
	blueCarOp.GeoM.Rotate(math.Pi)
	blueCarOp.GeoM.Translate(600, 280)
	screen.DrawImage(blueCar, blueCarOp)
}

func spawnYellowCar(screen *ebiten.Image) {
	yellowCarOp := &ebiten.DrawImageOptions{}
	// yellowCarOp.GeoM.Rotate(0)
	yellowCarOp.GeoM.Translate(0, 320)
	screen.DrawImage(yellowCar, yellowCarOp)
}

func spawnPurpleCar(screen *ebiten.Image) {
	purpleCarOp := &ebiten.DrawImageOptions{}
	purpleCarOp.GeoM.Rotate(3 * math.Pi / 2)
	purpleCarOp.GeoM.Translate(320, 600)
	screen.DrawImage(purpleCar, purpleCarOp)
}

func checkForUpdates() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetVehiclesDirectionsRequest{}
	res, err := serverClient.GetVehiclesDirections(ctx, req)
	if err != nil {
		log.Printf("Failed to get vehicle directions: %v", err)
		return
	}

	upSlice = res.Up
	downSlice = res.Down
	leftSlice = res.Left
	rightSlice = res.Right
}

func init() {
	var err error

	serverAddress = os.Getenv("SD_SERVER_ADDRESS")

	serverConn, err = grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	serverClient = pb.NewVehicleDiscoveryClient(serverConn)

	bg, _, err = ebitenutil.NewImageFromFile("./road.png")
	if err != nil {
		log.Fatal(err)
	}

	redCar, _, err = ebitenutil.NewImageFromFile("./red_car.png")
	if err != nil {
		log.Fatal(err)
	}

	blueCar, _, err = ebitenutil.NewImageFromFile("./blue_car.png")
	if err != nil {
		log.Fatal(err)
	}

	yellowCar, _, err = ebitenutil.NewImageFromFile("./yellow_car.png")
	if err != nil {
		log.Fatal(err)
	}

	purpleCar, _, err = ebitenutil.NewImageFromFile("./purple_car.png")
	if err != nil {
		log.Fatal(err)
	}
}
