import { Controller, Get, Param, ParseIntPipe } from '@nestjs/common';
import { VehicleLogService } from './vehicle_log.service';
import { Roles } from '../../authentication/auth/decorators/roles';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Role } from '@prisma/client';

@ApiTags('Vehicle Log')
@Controller('vehicle-log')
export class VehicleLogController {
  constructor(private readonly vehicleLogService: VehicleLogService) {}

  @Get('/:slotId')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getLogsBySlotId(@Param('slotId', ParseIntPipe) slotId: number) {
    const logs = await this.vehicleLogService.getLogsBySlotId(slotId);
    return {
      message: 'Lấy lịch sử xe thành công',
      data: {
        logs,
      },
    };
  }
}
