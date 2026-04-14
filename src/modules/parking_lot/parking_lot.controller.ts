import { Body, Controller, Get, Patch, Post, Param, ParseIntPipe } from '@nestjs/common';
import { ParkingLotService } from './parking_lot.service';
import { CreateParkingLotDTO, UpdateParkingLotDTO } from './dto';
import { Roles } from '../../authentication/auth/decorators/roles';
import { Role } from '@prisma/client';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';

@ApiTags('Parking Lots')
@Controller('parking-lots')
export class ParkingLotController {
  constructor(private readonly parkingLotService: ParkingLotService) {}

  @Post()
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async createParkingLot(@Body() dto: CreateParkingLotDTO) {
    const createdLot = await this.parkingLotService.createParkingLot(dto);
    return {
      message: 'Tạo bãi đỗ thành công',
      data: {
        parkingLot: createdLot
      }
    };
  }

  @Get()
  @ApiBearerAuth('access-token')
  async getAllParkingLot() {
    const parkingLots = await this.parkingLotService.getAllParkingLot();
    return {
      message: 'Lấy danh sách bãi đỗ thành công',
      data: {
        parkingLots
      }
    };
  }

  @Get('/:id')
  @ApiBearerAuth('access-token')
  async getParkingLotById(@Param('id', ParseIntPipe) id: number) {
    const parkingLot = await this.parkingLotService.getParkingLotById(id);
    return {
      message: 'Lấy thông tin bãi đỗ thành công',
      data: {
        parkingLot
      }
    };
  }

  @Patch('/:id')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async updateParkingLot(@Param('id', ParseIntPipe) id: number, @Body() dto: UpdateParkingLotDTO) {
    const updatedLot = await this.parkingLotService.updateParkingLot(id, dto);
    return {
      message: 'Cập nhật bãi đỗ thành công',
      data: {
        parkingLot: updatedLot
      }
    };
  }
}
