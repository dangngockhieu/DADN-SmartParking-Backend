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
import { JwtAuthGuard } from './authentication/auth/guards';

@Module({
  imports: [
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
    TokensModule
  ],
  providers: [
    {
      provide: APP_GUARD,
      useClass: JwtAuthGuard,
    }
  ]
})
export class AppModule {}
