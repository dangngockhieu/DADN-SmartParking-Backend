import { Controller, Post, Body } from '@nestjs/common';
import { IotDeviceService } from './iot_device.service';
import { CreateIoTDeviceDTO } from './dto';

@Controller('iot-devices')
export class IotDeviceController {
  constructor(private readonly iotDeviceService: IotDeviceService) {}

  @Post()
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
