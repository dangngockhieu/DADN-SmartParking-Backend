import { MailerService } from '@nestjs-modules/mailer';
import { Injectable, Logger, InternalServerErrorException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

@Injectable()
export class MailService {
  private readonly logger = new Logger(MailService.name);

  constructor(
    private mailerService: MailerService,
    private config: ConfigService
  ) {}

  async sendVerificationEmail(to: string, firstName: string, token: string) {
    try {
      const baseUrl = this.config.get<string>('VERIFY_BASE_URL');
      if (!baseUrl) {
        this.logger.error('Missing VERIFY_BASE_URL in environment variables');
        throw new InternalServerErrorException('Thiếu cấu hình VERIFY_BASE_URL');
      }

      const separator = baseUrl.includes('?') ? '&' : '?';
      const verifyUrl = `${baseUrl}${separator}token=${token}&email=${encodeURIComponent(to)}`;

      await this.mailerService.sendMail({
        to,
        subject: '[Smart Parking] Xác thực đăng ký tài khoản',
        template: './verification',
        context: { firstName, verifyUrl },
      });
      this.logger.log(`Verification email sent to ${to}`);
    } catch (error: any) {
      this.logger.error(`Error sending email to ${to}`, error.stack);
      throw new InternalServerErrorException('Gửi mail thất bại');
    }
  }

  async sendPasswordResetEmail(to: string, name: string, codeID: string) {
    try {
      await this.mailerService.sendMail({
        to,
        subject: '[Smart Parking] Mã xác nhận đặt lại mật khẩu',
        template: './reset-password',
        context: { name, codeID },
      });
      this.logger.log(`Reset email sent to ${to}`);
    } catch (error: any) {
      this.logger.error(`Error sending reset email to ${to}`, error.stack);
      throw new InternalServerErrorException('Gửi mail thất bại');
    }
  }
}