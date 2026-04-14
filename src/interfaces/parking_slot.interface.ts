import { SlotStatus } from "@prisma/client"

export interface UpdateParkingSlotDTO {
  changed: boolean,
  id: number,
  lot_id: number,
  name: string,
  message?: string,
  oldStatus: SlotStatus,
  newStatus:SlotStatus
}

export interface ParkingSlot {
  id: number,
  name: string,
  lotId: number,
  deviceMac: string,
  portNumber: number,
  status: SlotStatus,
  createdAt: Date,
  updatedAt: Date
}