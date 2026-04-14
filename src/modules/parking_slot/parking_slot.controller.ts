import { Body, Controller, Get, Param, ParseIntPipe, Patch, Post } from '@nestjs/common';
import { ParkingSlotService } from './parking_slot.service';
import { AdminUpdateParkingLotDTO, ChangeSlotDeviceDTO, CreateParkingSlotDTO, SensorUpdateParkingLotDTO } from './dto';
import { Roles } from '../../authentication/auth/decorators/roles';
import { Role } from '@prisma/client';
import { Public } from '../../authentication/auth/decorators/customize';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';

@ApiTags('Parking Slots')
@Controller('parking-slots')
export class ParkingSlotController {
  constructor(private readonly parkingSlotService: ParkingSlotService) {}

  // Create parking slot
  @Post()
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async createParkingSlot(@Body() dto: CreateParkingSlotDTO) {
    const newSlot = await this.parkingSlotService.createParkingSlot(dto);
    return {
      message: 'Tạo vị trí đỗ thành công',
      data: {
        parkingSlot: newSlot
      }
    };
  }

  // Get parking slot by id
  @Get('/:id')
  @ApiBearerAuth('access-token')
  async getParkingSlotById(@Param('id', ParseIntPipe) id: number) {
    const slot = await this.parkingSlotService.getSlotById(id);
    return {
      message: 'Lấy thông tin vị trí đỗ thành công',
      data: {
        parkingSlot: slot
      }
    };
  }

  // Admin update parking slot status
  @Patch('/admin/:id')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async adminUpdateParkingSlotStatus(@Param('id', ParseIntPipe) id: number, @Body() dto: AdminUpdateParkingLotDTO) {
    const updatedSlot = await this.parkingSlotService.adminUpdateParkingSlotStatus(id, dto);
    return {
      message: updatedSlot.message,
      data: {
        parkingSlot: updatedSlot
      }
    };
  }

  // Sensor update parking slot status
  @Post('/sensor')
  @Public()
  async sensorUpdateParkingSlotStatus(@Body() dto: SensorUpdateParkingLotDTO) {
    const updatedSlot = await this.parkingSlotService.sensorUpdateParkingSlotStatus(dto);
    return {
      message: updatedSlot.message,
      data: {
        parkingSlot: updatedSlot
      }
    };
  }

  // Change slot device
  @Patch('/:id/device')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async changeSlotDevice(@Param('id', ParseIntPipe) id: number, @Body() dto: ChangeSlotDeviceDTO) {
    await this.parkingSlotService.changeDevice(id, dto);
    return {
      message: 'Cập nhật thiết bị thành công'
    };
  }
}
