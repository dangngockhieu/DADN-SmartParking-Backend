import { Module } from '@nestjs/common';
import { IotDeviceService } from './iot_device.service';
import { IotDeviceController } from './iot_device.controller';

@Module({
  controllers: [IotDeviceController],
  providers: [IotDeviceService]
})
export class IotDeviceModule {}
