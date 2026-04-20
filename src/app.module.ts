import { Module } from '@nestjs/common';
import { AuthModule } from './authentication/auth/auth.module';
import { MailModule } from './authentication/mail/mail.module';
import { TokensModule } from './authentication/tokens/tokens.module';
import { ConfigModule } from '@nestjs/config';
import { ServeStaticModule } from '@nestjs/serve-static';
import { join } from 'path';
import { PrismaModule } from './prisma/prisma.module';
import { ParkingLotModule } from './modules/parking_lot/parking_lot.module';
import { ParkingSlotModule } from './modules/parking_slot/parking_slot.module';
import { UserModule } from './modules/user/user.module';
import { APP_GUARD } from '@nestjs/core';
import { JwtAuthGuard } from './common/guards';
import { SlotHistoryModule } from './modules/slot_history/slot_history.module';
import { IotDeviceModule } from './modules/iot_device/iot_device.module';
import { StatisticsModule } from './modules/statistics/statistics.module';
import { ThrottlerGuard, ThrottlerModule } from '@nestjs/throttler';

@Module({
  imports: [
    ThrottlerModule.forRoot([
      {
        ttl: 60000,
        limit: 10,
      }
    ]),
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    ServeStaticModule.forRoot({
      rootPath: join(__dirname, '..', 'public'),
    }),
    PrismaModule,
    ParkingLotModule,
    ParkingSlotModule,
    UserModule,
    AuthModule,
    MailModule,
    TokensModule,
    SlotHistoryModule,
    IotDeviceModule,
    StatisticsModule,
  ],
  providers: [
    {
      provide: APP_GUARD,
      useClass: JwtAuthGuard,
    },
    {
      provide: APP_GUARD,
      useClass: ThrottlerGuard
    }
  ]
})
export class AppModule {}
