import { Module } from '@nestjs/common';
import { MailerModule } from '@nestjs-modules/mailer';
import { HandlebarsAdapter } from '@nestjs-modules/mailer/adapters/handlebars.adapter';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { join } from 'path';
import { existsSync } from 'fs';
import { MailService } from './mail.service';

const resolveTemplateDir = () => {
  const candidates = [
    join(process.cwd(), 'dist', 'authentication', 'mail', 'templates'),
    join(process.cwd(), 'dist', 'templates'),
    join(process.cwd(), 'src', 'authentication', 'mail', 'templates'),
    join(__dirname, 'templates'),
  ];

  const matched = candidates.find((dir) => existsSync(dir));
  return matched ?? candidates[candidates.length - 1];
};

@Module({
  imports: [
    MailerModule.forRootAsync({
      imports: [ConfigModule],
      useFactory: async (config: ConfigService) => ({
        transport: {
          host: 'smtp.gmail.com',
          secure: false, // true cho port 465, false cho các port khác
          auth: {
            user: config.get('MAIL_USER'),
            pass: config.get('MAIL_PASS'),
          },
        },
        template: {
          dir: resolveTemplateDir(),
          adapter: new HandlebarsAdapter(), // Khởi tạo tại đây
          options: {
            strict: true,
          },
        },
      }),
      inject: [ConfigService],
    }),
  ],
  providers: [MailService],
  exports: [MailService],
})
export class MailModule {}