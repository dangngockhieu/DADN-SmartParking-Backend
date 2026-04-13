import { Module } from '@nestjs/common';
import { AuthService } from './auth.service';
import { AuthController } from './auth.controller';
import { UserModule } from '../../modules/user/user.module';
import { MailModule } from '../mail/mail.module';
import { TokensModule } from '../tokens/tokens.module';
import { JwtStrategy, LocalStrategy } from './strategies';

@Module({
  imports: [UserModule, MailModule, TokensModule],
  controllers: [AuthController],
  providers: [AuthService, JwtStrategy, LocalStrategy],
  exports: [AuthService],
})
export class AuthModule {}
