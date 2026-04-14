import { Controller, Post, Body } from '@nestjs/common';
import { IotDeviceService } from './iot_device.service';
import { CreateIoTDeviceDTO } from './dto';
import { Roles } from '../../authentication/auth/decorators/roles';
import { Role } from '@prisma/client';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';

@ApiTags('IoT Devices')
@Controller('iot-devices')
export class IotDeviceController {
  constructor(private readonly iotDeviceService: IotDeviceService) {}

  @Post()
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async create(@Body() dto: CreateIoTDeviceDTO) {
    const device = await this.iotDeviceService.createDevice(dto);
    return {
      message: 'Tạo thiết bị IoT thành công',
      data: {
        device
      }
    };
  }
}
