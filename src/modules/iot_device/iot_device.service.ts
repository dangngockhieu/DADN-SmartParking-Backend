import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';
import { CreateIoTDeviceDTO } from './dto';
import { BadRequestException } from '../../common/exception';
@Injectable()
export class IotDeviceService {
  constructor(private prisma: PrismaService) {}

  // Create new IoT device
  async createDevice(dto: CreateIoTDeviceDTO) {
    const exist = await this.prisma.ioTDevice.findUnique({
      where: { macAddress: dto.macAddress },
    });

    if (exist) {
      throw new BadRequestException('Device already exists');
    }

    const device = await this.prisma.ioTDevice.create({
      data: {
        macAddress: dto.macAddress,
        deviceName: dto.deviceName,
        lotId: dto.lotId,
        status: 'ACTIVE',
        lastSeen: new Date(),
      },
    });

    return device;
  }
}
