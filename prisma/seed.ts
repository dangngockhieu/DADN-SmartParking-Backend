import { PrismaClient, Role, SlotStatus, DeviceStatus } from '@prisma/client';
import * as argon from 'argon2';

const prisma = new PrismaClient();

async function main() {
  // 1. Create admin user
  const hashedPassword = await argon.hash('123456');

  const admin = await prisma.user.upsert({
    where: { email: 'admin@gmail.com' },
    update: {},
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

  console.log('✅ Admin created:', admin.email);

  // 2. Create Parking Lot A
  const lotA = await prisma.parkingLot.upsert({
    where: { id: 1 },
    update: {},
    create: {
      name: 'A',
      location: 'Main Area',
    },
  });

  console.log('✅ Parking lot created:', lotA.name);

  // 3. Create IoT Device (cần để gắn slot)
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

  console.log('✅ Device created:', device.macAddress);

  // 4. Create 8 slots
  const slots = [];

  for (let i = 1; i <= 8; i++) {
    const slot = await prisma.parkingSlot.upsert({
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

    slots.push(slot);
  }

  console.log(`✅ Created ${slots.length} slots for lot A`);
}

main()
  .then(async () => {
    await prisma.$disconnect();
  })
  .catch(async (e) => {
    console.error(e);
    await prisma.$disconnect();
    process.exit(1);
  });