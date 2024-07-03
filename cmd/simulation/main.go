package main

import (
	"context"

	// "fmt"
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

	upSlice    = []*Car{}
	downSlice  = []*Car{}
	rightSlice = []*Car{}
	leftSlice  = []*Car{}
)

type Direction int

const (
	up Direction = iota
	down
	left
	right
)

const (
	width  = 600
	height = 600

	// UpCenter   = 320
	// DownCenter = 600
)

type CarRepresentation struct {
	Image        *ebiten.Image
	ImageOptions *ebiten.DrawImageOptions
}

type Car struct {
	Vehicle        *pb.Vehicle
	Representation *CarRepresentation
	ShouldWalk     bool
	Dir            Direction
	Position       []float64
	Spawned        bool
}

func (c *Car) Spawn(screen *ebiten.Image) {
	// log.Println("Spawning car...")
	if !c.Spawned {
		switch c.Dir {
		case up:
			c.Representation.Image = redCar
			c.Representation.ImageOptions.GeoM.Rotate(math.Pi / 2)
			c.Representation.ImageOptions.GeoM.Translate(280, 0)
		case down:
			c.Representation.Image = blueCar
			c.Representation.ImageOptions.GeoM.Rotate(math.Pi)
			c.Representation.ImageOptions.GeoM.Translate(600, 280)
		case left:
			c.Representation.Image = yellowCar
			c.Representation.ImageOptions.GeoM.Translate(0, 320)
		case right:
			c.Representation.Image = purpleCar
			c.Representation.ImageOptions.GeoM.Rotate(3 * math.Pi / 2)
			c.Representation.ImageOptions.GeoM.Translate(320, 600)
		}
	}

	screen.DrawImage(c.Representation.Image, c.Representation.ImageOptions)

	c.Spawned = true
	// }
}

func (c *Car) Walk() {
	if c.ShouldWalk {
		log.Printf("car %v should walk", c.Vehicle.Address)
		switch c.Dir {
		case up:
			c.Position[0] += 30
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case down:
			c.Position[0] -= 30
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case left:
			c.Position[1] += 30
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case right:
			c.Position[1] -= 30
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		}

		c.ShouldWalk = false
	}
}

// Game implements ebiten.Game interface.
type Game struct{}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	checkForUpdates()
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	bgOp := &ebiten.DrawImageOptions{}
	screen.DrawImage(bg, bgOp)
	// screen.DrawImage(redCar, nil)

	// spawn all non-spawned Cars
	// log.Printf("rightSlice len == %d", len(rightSlice))
	// for _, r := range rightSlice {
	// 	r.Spawn(screen)
	//
	// }

	for _, arr := range [][]*Car{upSlice, downSlice, leftSlice, rightSlice} {
		for _, v := range arr {
			// log.Println("Spawning car...")
			v.Spawn(screen)
			v.Walk()
		}
	}

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
	ebiten.SetWindowTitle("Intersection simulator")
	log.Println("Starting intersection simulator...")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func contains(v *pb.Vehicle, vArr []*Car) (bool, int) {
	for i, r := range vArr {
		if r.Vehicle.Id == v.Id {
			return true, i
		}
	}

	return false, -1
}

func updateCarArr(dSlice []*Car, res []*pb.Vehicle, dir string) (resultArray []*Car) {
	var dT Direction

	switch dir {
	case "up":
		dT = up
	case "down":
		dT = down
	case "left":
		dT = left
	case "right":
		dT = right
	}

	log.Println(res)

	for _, r := range res {
		if o, idx := contains(r, dSlice); o {
			// log.Println("Car already in slice")
			// log.Println("!dSlice[idx].Vehicle.ShouldWalk", !dSlice[idx].Vehicle.ShouldWalk)
			// log.Println("r.ShouldWalk", r.ShouldWalk)
			// log.Println("!dSlice[idx].Vehicle.ShouldWalk && r.ShouldWalk", !dSlice[idx].Vehicle.ShouldWalk && r.ShouldWalk)

			if !dSlice[idx].Vehicle.ShouldWalk && r.ShouldWalk {
				dSlice[idx].ShouldWalk = true
				log.Println("This car should walk")
			}

		} else {
			// log.Printf("adding car on direction %d", dT)
			dSlice = append(dSlice, &Car{
				Vehicle: r,
				Representation: &CarRepresentation{
					Image:        nil,
					ImageOptions: &ebiten.DrawImageOptions{},
				},
				ShouldWalk: false,
				Dir:        dT,
				Spawned:    false,
				Position:   []float64{},
			})
		}
	}
	return dSlice
}

func checkForUpdates() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// log.Println("Requiring directions from server...")
	req := &pb.GetVehiclesDirectionsRequest{
		RequesterName: "simulation_ui",
	}
	res, err := serverClient.GetVehiclesDirections(ctx, req)
	if err != nil {
		log.Printf("Failed to get vehicle directions: %v", err)
		return
	}

	upSlice = updateCarArr(upSlice, res.Up, "up")
	downSlice = updateCarArr(downSlice, res.Down, "down")
	leftSlice = updateCarArr(leftSlice, res.Left, "left")
	rightSlice = updateCarArr(rightSlice, res.Right, "right")

}

func init() {
	var err error

	serverAddress = os.Getenv("SD_SERVER_ADDRESS")
	if serverAddress == "" {
		log.Fatalf("No server address provided. Set env variable SD_SERVER_ADDRESS")
	}

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
