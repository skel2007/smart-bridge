package tuya

import (
	"context"

	"github.com/skel2007/smart-bridge/internal/cache"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
)

type cachedSpecificationsAPI struct {
	next           cloudAPI
	specifications *cache.MemoizedGetter[cloud.DeviceSpecifications]
}

func newCachedSpecificationsAPI(next cloudAPI) cloudAPI {
	return &cachedSpecificationsAPI{
		next:           next,
		specifications: cache.NewMemoizedGetter(next.GetDeviceSpecifications),
	}
}

func (api *cachedSpecificationsAPI) ListProjectDevices(ctx context.Context) ([]cloud.Device, error) {
	return api.next.ListProjectDevices(ctx)
}

func (api *cachedSpecificationsAPI) GetDeviceSpecifications(ctx context.Context, deviceID string) (cloud.DeviceSpecifications, error) {
	return api.specifications.Get(ctx, deviceID)
}

func (api *cachedSpecificationsAPI) GetDeviceStatus(ctx context.Context, deviceID string) ([]cloud.DeviceStatus, error) {
	return api.next.GetDeviceStatus(ctx, deviceID)
}

func (api *cachedSpecificationsAPI) SendCommands(ctx context.Context, deviceID string, commands []cloud.Command) error {
	return api.next.SendCommands(ctx, deviceID, commands)
}
