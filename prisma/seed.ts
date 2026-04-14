import { NestFactory } from '@nestjs/core';
import { AppModule } from '../src/app.module';
import { PrismaService } from '../src/prisma/prisma.service';
import { SlotStatus, DeviceStatus, Role } from '@prisma/client';
import * as argon from 'argon2';

async function seed(prisma: PrismaService) {
  // 1. Admin
  const hashedPassword = await argon.hash('123456');

  const admin = await prisma.user.upsert({
    where: { email: 'admin@gmail.com' },
    update: {}, // có rồi thì bỏ qua
    create: {
      firstName: 'Admin',
      lastName: 'System',
      email: 'admin@gmail.com',
      username: 'admin',
      password: hashedPassword,
      role: Role.ADMIN,
      isVerified: true,
    },
  });

  console.log('✅ Admin:', admin.email);

  // 2. Parking Lot A
  const lotA = await prisma.parkingLot.upsert({
    where: { id: 1 }, // hoặc bạn có thể dùng unique name nếu có
    update: {},
    create: {
      name: 'A',
      location: 'Main Area',
    },
  });

  console.log('✅ Lot:', lotA.name);

  // 3. IoT Device
  const device = await prisma.ioTDevice.upsert({
    where: { macAddress: 'DEVICE_A_001' },
    update: {},
    create: {
      macAddress: 'DEVICE_A_001',
      deviceName: 'Gate Controller A',
      status: DeviceStatus.ACTIVE,
      lotId: lotA.id,
    },
  });

  console.log('✅ Device:', device.macAddress);

  // 4. 8 Parking Slots
  for (let i = 1; i <= 8; i++) {
    await prisma.parkingSlot.upsert({
      where: {
        deviceMac_portNumber: {
          deviceMac: device.macAddress,
          portNumber: i,
        },
      },
      update: {},
      create: {
        name: `A${i}`,
        lotId: lotA.id,
        deviceMac: device.macAddress,
        portNumber: i,
        status: SlotStatus.AVAILABLE,
      },
    });
  }

  console.log('✅ 8 slots created');
}

async function bootstrap() {
  const app = await NestFactory.createApplicationContext(AppModule);

  const prisma = app.get(PrismaService);

  try {
    await seed(prisma);
  } catch (e) {
    console.error('❌ Seed error:', e);
  } finally {
    await app.close();
  }
}

bootstrap();