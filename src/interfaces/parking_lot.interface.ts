import { SlotStatus } from "@prisma/client"

export interface ParkingLotAll {
  id: number,
  name: string,
  location: string
}

export interface Parking_Slot {
  id: number,
  name: string,
  status: SlotStatus,
  deviceMac: string,
  portNumber: number
}

export interface ParkingLot {
  id: number,
  name: string,
  location: string,
  slots: Parking_Slot[]
}

export interface ParkingLotStats {
  total: number,
  available: number,
  occupied: number,
  maintain: number
}

export interface ParkingLotWithStats extends ParkingLot {
  stats: ParkingLotStats
}