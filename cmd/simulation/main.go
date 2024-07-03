package main

import (
	"context"
	"fmt"

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

	deleted = []*Car{}
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
			c.Representation.Image, _, _ = ebitenutil.NewImageFromFile("./red_car.png")
			c.Position = []float64{280, 0}
			c.Representation.ImageOptions.GeoM.Rotate(math.Pi / 2)
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case down:
			c.Representation.Image, _, _ = ebitenutil.NewImageFromFile("./blue_car.png")
			c.Position = []float64{320, 600}
			c.Representation.ImageOptions.GeoM.Rotate(3 * math.Pi / 2)
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case left:
			c.Representation.Image, _, _ = ebitenutil.NewImageFromFile("./yellow_car.png")
			c.Position = []float64{0, 320}
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		case right:
			c.Representation.Image, _, _ = ebitenutil.NewImageFromFile("./purple_car.png")
			c.Position = []float64{600, 280}
			c.Representation.ImageOptions.GeoM.Rotate(math.Pi)
			c.Representation.ImageOptions.GeoM.Translate(c.Position[0], c.Position[1])
		}
		screen.DrawImage(c.Representation.Image, c.Representation.ImageOptions)

		c.Spawned = true
	}
}

func allowLast(arr []*Car) []*Car {
	for i := 0; i < len(arr); i++ {
		arr[i].ShouldWalk = true
	}

	return arr
}

func (c *Car) Walk(screen *ebiten.Image, idx int) {
	if c.ShouldWalk && c.Spawned {
		switch c.Dir {

		case up:
			c.Position[1] += 30
			c.Representation.ImageOptions.GeoM.Translate(0, 30)
			fmt.Println("c.Position[1]: ", c.Position[1])
			if c.Position[1] >= 300 {
				c.Representation.Image.Clear()
				if len(upSlice) > 0 {
					upSlice = append(upSlice[:idx], upSlice[idx+1:]...)
					upSlice = allowLast(upSlice)
				} else {
					upSlice = []*Car{}
				}
				deleted = append(deleted, c)
			}
		case down:
			c.Position[1] += -30
			// c.Position[1] -= 1
			fmt.Println("c.Position[1]: ", c.Position[1])
			c.Representation.ImageOptions.GeoM.Translate(0, -30)
			if c.Position[1] <= 300 {
				c.Representation.Image.Clear()
				if len(downSlice) > 0 {
					downSlice = append(downSlice[:idx], downSlice[idx+1:]...)
					downSlice = allowLast(downSlice)
				} else {
					downSlice = []*Car{}
				}
				deleted = append(deleted, c)
			}
		case left:
			c.Position[0] += 30
			c.Representation.ImageOptions.GeoM.Translate(30, 0)
			if c.Position[0] >= 300 {
				c.Representation.Image.Clear()
				if len(leftSlice) > 0 {
					leftSlice = append(leftSlice[:idx], leftSlice[idx+1:]...)
					leftSlice = allowLast(leftSlice)
				} else {
					leftSlice = []*Car{}
				}
				deleted = append(deleted, c)
			}
		case right:
			c.Position[0] += -30
			c.Representation.ImageOptions.GeoM.Translate(-30, 0)
			if c.Position[0] <= 300 {
				c.Representation.Image.Clear()
				if len(rightSlice) > 0 {
					rightSlice = append(rightSlice[:idx], rightSlice[idx+1:]...)
					rightSlice = allowLast(rightSlice)
				} else {
					rightSlice = []*Car{}
				}
				deleted = append(deleted, c)
			}
		}
		c.ShouldWalk = false
	}
	if c.Spawned {
		screen.DrawImage(c.Representation.Image, c.Representation.ImageOptions)
		time.Sleep(250 * time.Millisecond)
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
		if len(arr) == 1 {
			arr = allowLast(arr)
		}
		for idx, v := range arr {
			// log.Println("Spawning car...")
			v.Walk(screen, idx)
			// time.Sleep(500 * time.Millisecond)
			v.Spawn(screen)
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

	for _, r := range res {
		// checks if r.shouldwalk
		// log.Println("r.ShouldWalk:", r.ShouldWalk)

		if o, _ := contains(r, deleted); o {
			log.Println("this car was removed")
			continue
		}

		if o, idx := contains(r, dSlice); o {
			if !dSlice[idx].Vehicle.ShouldWalk && r.ShouldWalk {
				dSlice[idx].ShouldWalk = true
			}

		} else {
			log.Println("Adding car")
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

	log.Println("got response: ", res)

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
