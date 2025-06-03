package externals

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

type AllAppExternals struct {
	All []BaseExternal // Array of external dependencies
}

type BaseExternal interface {
	ConnectRaw() error      // Connect using External Connect()
	Healthcheck() error     // Implement healthcheck logic
	SuccessMessage() string // Success message upon connection
}

type External[T any] interface {
	BaseExternal
	Connect() (T, error) // Implement connection logic here for external dependency
}

// Register all external dependencies
func RegisterExternals(allExternals []BaseExternal) (*AllAppExternals, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, _ := errgroup.WithContext(ctx)

	// Registering dynamically
	for _, ext := range allExternals {
		ext := ext
		g.Go(func() error {
			if err := ext.ConnectRaw(); err != nil {
				return err
			}

			if err := ext.Healthcheck(); err != nil {
				return err
			}

			log.Println(ext.SuccessMessage())

			return nil
		})
	}

	return &AllAppExternals{
		All: allExternals,
	}, g.Wait()
}

// Pick one external from list of registered externals by type pointer
func GetExternal[T BaseExternal](externals *AllAppExternals) (T, error) {
	var zero T
	for _, ext := range externals.All {
		if t, ok := ext.(T); ok {
			return t, nil
		}
	}
	return zero, fmt.Errorf("external of type %T not found", zero)
}
