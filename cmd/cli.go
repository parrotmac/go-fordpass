package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/manifoldco/promptui"

	fordpass "github.com/parrotmac/go-fordpass"
)

const (
	keyUsername = "Username"
	keyPassword = "Password"
	keyVIN      = "VIN"

	actionGetStatus   = "Get Status"
	actionLockDoors   = "Lock Doors"
	actionUnlockDoors = "Unlock Doors"
	actionStartEngine = "Start Engine"
	actionStopEngine  = "Stop Engine"
)

func mustGetParams() (string, string, string) {
	requiredParams := map[string]string{
		keyUsername: os.Getenv("FORD_USERNAME"),
		keyPassword: os.Getenv("FORD_PASSWORD"),
		keyVIN:      os.Getenv("VEHICLE_VIN"),
	}

	for paramKey, paramValue := range requiredParams {
		if paramValue == "" {
			requiredParams[paramKey] = mustPromptString(paramKey)
		}
	}

	username := requiredParams[keyUsername]
	password := requiredParams[keyPassword]
	vin := requiredParams[keyVIN]
	return username, password, vin
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	username, password, vin := mustGetParams()
	vehicleAPI := fordpass.NewVehicleAPI(username, password, vin)

	cliActions := []string{
		actionGetStatus,
		actionLockDoors,
		actionUnlockDoors,
		actionStartEngine,
		actionStopEngine,
	}

	switch mustSelect("Choose Action", cliActions) {
	case actionGetStatus:
		statusInfo, err := vehicleAPI.Status(ctx)
		musntErr("Get vehicle status", err)
		log.Printf("%+v", statusInfo)
	case actionLockDoors:
		err := vehicleAPI.Lock(ctx)
		musntErr("Lock doors", err)
	case actionUnlockDoors:
		err := vehicleAPI.Unlock(ctx)
		musntErr("Unlock doors", err)
	case actionStartEngine:
		err := vehicleAPI.StartEngine(ctx)
		musntErr("Start engine", err)
	case actionStopEngine:
		err := vehicleAPI.StopEngine(ctx)
		musntErr("Stop engine", err)

	default:
		log.Fatalln("Failed to select an available action.")
	}

}

func musntErr(action string, err error) {
	if err != nil {
		log.Fatalf("%s failed: %+v", action, err)
	}
}

func mustPromptString(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}
	result, err := prompt.Run()
	musntErr("Prompt for "+label, err)
	return result
}

func mustSelect(label string, options []string) string {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}

	_, result, err := prompt.Run()
	musntErr("Prompt for "+label, err)
	return result
}
