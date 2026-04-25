package iot_device

type CreateIoTDeviceRequest struct {
	MacAddress string `json:"mac_address" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
	LotID      *uint  `json:"lot_id"`
}
