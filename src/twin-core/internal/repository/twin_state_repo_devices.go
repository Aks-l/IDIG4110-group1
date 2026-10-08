package repository

import (
	"context"
	"fmt"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Row level security scopes devices to app.home_id, WithHome sets it
const listDevicesByHomeQuery = `
	SELECT id::text, home_id::text, area_id::text, gateway_id::text,
		external_id, name, manufacturer, model, sw_version, device_type
	FROM devices
	ORDER BY name
`

// Lists the devices of one home
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Devices of the home
//   - Error on failure
func (r *TwinStateRepoImpl) ListDevicesByHome(ctx context.Context, homeID string) ([]domain.Device, error) {
	devices := []domain.Device{}
	err := r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listDevicesByHomeQuery)
		if err != nil {
			return fmt.Errorf("list devices: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var d domain.Device
			if err := rows.Scan(&d.ID, &d.HomeID, &d.AreaID, &d.GatewayID, &d.ExternalID, &d.Name, &d.Manufacturer, &d.Model, &d.SwVersion, &d.DeviceType); err != nil {
				return fmt.Errorf("list devices: %w", err)
			}
			devices = append(devices, d)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return devices, nil
}